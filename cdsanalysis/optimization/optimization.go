package optimization

import (
	"fmt"

	errors "github.com/edgelaboratories/go-errors/goerror"
	"gonum.org/v1/gonum/optimize"
)

type ObjectiveFunction interface {
	Value(x []float64) float64
	Gradient(x []float64) []float64
}

type BoxedProblem struct {
	LowerBounds  []float64
	UpperBounds  []float64
	InitialGuess []float64
	Objective    ObjectiveFunction
}

type Settings struct {
	maxOutterIterations int
}

func NewSettings(maxOutterIter int) Settings {
	return Settings{maxOutterIter}
}

type SubProblem struct {
	globalProblem     BoxedProblem
	activeLowerBounds []bool
	activeUpperBounds []bool
	activeCoordinates []int
}

func (c SubProblem) activeToGlobal(x []float64) []float64 {
	nbVar := len(c.globalProblem.InitialGuess)
	xglobal := make([]float64, nbVar)

	iActive := 0

	for i := range nbVar {
		switch {
		case c.activeLowerBounds[i]:
			xglobal[i] = c.globalProblem.LowerBounds[i]
		case c.activeUpperBounds[i]:
			xglobal[i] = c.globalProblem.UpperBounds[i]
		default:
			xglobal[c.activeCoordinates[iActive]] = x[iActive]
			iActive++
		}
	}

	return xglobal
}

func (c SubProblem) Value(x []float64) float64 {
	return c.globalProblem.Objective.Value(c.activeToGlobal(x))
}

func (c SubProblem) Gradient(grad []float64, x []float64) {
	g := c.globalProblem.Objective.Gradient(c.activeToGlobal(x))

	for i := range c.activeCoordinates {
		grad[i] = g[c.activeCoordinates[i]]
	}
}

type Solver struct {
	settings Settings
}

func NewSolver(settings Settings) Solver {
	return Solver{settings}
}

func (s Solver) Minimize(problem BoxedProblem) ([]float64, error) {
	if err := s.checkProblemValidity(problem); err != nil {
		return nil, err
	}

	nbVar := len(problem.InitialGuess)

	currentGuess := append([]float64{}, problem.InitialGuess...)
	activeLowerBounds := make([]bool, nbVar) // No active lower bounds to start
	activeUpperBounds := make([]bool, nbVar) // No active upper bounds to start

	for range s.settings.maxOutterIterations {
		// First step is to build the problem restricted
		// using the active bounds, for which the optimization
		// will be performed
		activeCoordinates := make([]int, 0, nbVar)
		restrictedGuess := make([]float64, 0, nbVar)

		for i := range nbVar {
			switch {
			case activeLowerBounds[i]:
				currentGuess[i] = problem.LowerBounds[i]
			case activeUpperBounds[i]:
				currentGuess[i] = problem.UpperBounds[i]
			default:
				activeCoordinates = append(activeCoordinates, i)
				restrictedGuess = append(restrictedGuess, currentGuess[i])
			}
		}

		// This is the next iteration values
		// there are two ways to get it:
		// 1. if there are no unknown, just use the active bounds
		// 2. if there are unknowns, minimize
		fullResult := make([]float64, nbVar)

		if len(activeCoordinates) == 0 {
			for i := range nbVar {
				switch {
				case activeLowerBounds[i]:
					fullResult[i] = problem.LowerBounds[i]
				case activeUpperBounds[i]:
					fullResult[i] = problem.UpperBounds[i]
				default:
					return nil, errors.Bug("inconsistent internal state: no unknown but no bound")
				}
			}
		} else {
			subProblem := SubProblem{
				problem, activeLowerBounds, activeUpperBounds, activeCoordinates,
			}

			p := optimize.Problem{
				Func: subProblem.Value,
			}

			// Solve the sub problem without taking into account
			// the constraints which are inactive at the moment.

			settings := optimize.Settings{
				Converger: &optimize.FunctionConverge{
					Absolute:   1e-14,
					Relative:   1e-2,
					Iterations: 10,
				},
			}

			result, err := optimize.Minimize(p, restrictedGuess, &settings, nil)
			if err != nil {
				return nil, fmt.Errorf("internal minimization globally failed: %w", err)
			}

			if err = result.Status.Err(); err != nil {
				return nil, fmt.Errorf("internal minimization failed: %w", err)
			}

			fullResult = subProblem.activeToGlobal(result.X)
		}

		// Check whether the new guess from the sub problem
		// is valid or if it breaches some constraints that
		// were inactive.
		// We keep in memory the first constraint that was
		// breached while going from the previous guess to
		// the optimum of the sub problem.

		foundBrokenConstraint := false
		lambdaMin := 1.0
		isLowerConstraint := true

		var constraintID int

		for i := range nbVar {
			if !activeLowerBounds[i] && fullResult[i] <= problem.LowerBounds[i] {
				lambda := (problem.LowerBounds[i] - currentGuess[i]) / (fullResult[i] - currentGuess[i])
				if lambda <= lambdaMin {
					lambdaMin = lambda
					constraintID = i
					isLowerConstraint = true
					foundBrokenConstraint = true
				}
			} else if !activeUpperBounds[i] && fullResult[i] >= problem.UpperBounds[i] {
				lambda := (problem.UpperBounds[i] - currentGuess[i]) / (fullResult[i] - currentGuess[i])
				if lambda <= lambdaMin {
					lambdaMin = lambda
					constraintID = i
					isLowerConstraint = false
					foundBrokenConstraint = true
				}
			}
		}

		// Set the first constraint breached
		// as active for the next iteration.
		if foundBrokenConstraint {
			if isLowerConstraint {
				activeLowerBounds[constraintID] = true
			} else {
				activeUpperBounds[constraintID] = true
			}
		}

		// Build a new current guess as the first time
		// we breached the constraints.
		for i := range nbVar {
			currentGuess[i] += lambdaMin * (fullResult[i] - currentGuess[i])
		}

		// Recheck all the active constraints
		// and release the one with the most negative gradient
		// in the normal direction
		// (which corresponds to checking the sign of the
		// Lagrange multiplier):
		// - for a lower bound, we expect the derivative
		//    to be positive at optimum
		// - for an upper bound, we expect the derivative
		//    to be negative at optimum

		g := problem.Objective.Gradient(currentGuess)

		multiplierMin := 0.0
		isLowerMultiplier := true

		var multiplierID int

		for i := range nbVar {
			if activeLowerBounds[i] && g[i] < multiplierMin {
				multiplierMin = g[i]
				multiplierID = i
				isLowerMultiplier = true
			} else if activeUpperBounds[i] && -g[i] < multiplierMin {
				multiplierMin = -g[i]
				multiplierID = i
				isLowerMultiplier = false
			}
		}

		if multiplierMin < 0.0 {
			if isLowerMultiplier {
				activeLowerBounds[multiplierID] = true
			} else {
				activeUpperBounds[multiplierID] = true
			}
		} else if !foundBrokenConstraint {
			// In this case:
			// - The solution is optimal given the current active constraints
			// - There were no broken constraints
			// - Gradient has the correct sign normal to the constraints
			// ==> Optimal for the original problem
			return currentGuess, nil
		}
	}

	return nil, errors.Bug("optimization failed to converge in %d iterations", s.settings.maxOutterIterations)
}

type validityError string

func (e validityError) Error() string {
	return string(e)
}

const (
	errNoInitialGuess         = validityError("no initial guess defined")
	errLowerBoundCount        = validityError("lower bounds size does not match the problem")
	errUpperBoundCount        = validityError("upper bounds size does not match the problem")
	errIncompatibleBounds     = validityError("incompatible bounds")
	errInitiaGuessOutOfDomain = validityError("initial guess out of the domain")
)

func (s Solver) checkProblemValidity(problem BoxedProblem) error {
	nbVar := len(problem.InitialGuess)

	if nbVar == 0 {
		return errNoInitialGuess
	}

	if nbVar != len(problem.LowerBounds) {
		return errLowerBoundCount
	}

	if nbVar != len(problem.UpperBounds) {
		return errUpperBoundCount
	}

	for i := range nbVar {
		if problem.LowerBounds[i] >= problem.UpperBounds[i] {
			return errIncompatibleBounds
		}

		if problem.LowerBounds[i] > problem.InitialGuess[i] {
			return errInitiaGuessOutOfDomain
		}

		if problem.UpperBounds[i] < problem.InitialGuess[i] {
			return errInitiaGuessOutOfDomain
		}
	}

	return nil
}
