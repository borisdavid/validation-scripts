package integration

// interval represents an interval (a,b) on the real line.
type interval struct {
	a float64
	b float64
}

// Length returns the length of the interval.
func (i interval) length() float64 {
	return i.b - i.a
}

// midPoint returns the mid point of the interval.
func (i interval) midPoint() float64 {
	return 0.5 * (i.a + i.b)
}

func (i interval) halve() []*interval {
	mid := i.midPoint()

	return []*interval{
		{a: i.a, b: mid},
		{a: mid, b: i.b},
	}
}
