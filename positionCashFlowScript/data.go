package main

type inputData struct {
	id       string
	currency string
	quantity float64
}

var data = []inputData{
	{"26e24f77-06d2-4408-966a-3a3c06064ac6", "EUR", 1.0},
	{"4b382cdc-9097-4886-8127-63130e313429", "HKD", 1.0},
	{"aa375a52-df66-4aba-beb6-ceed90187ac4", "GBP", 1.0},
	{"26a68fb2-007a-481c-8678-b62e861605b6", "CHF", 1.0},

	{"b8fd4728-a286-4b92-a85a-59a15243ccbd", "INR", 1.0},

	/* // Bonds.
	{"1cc033cd-076f-4b34-b477-0a9faf89d932", "EUR", 1.0},
	{"7eeb39da-fb50-4f6c-814f-59701e975113", "GBP", 1.0},
	{"db88dbc1-ece3-4658-815b-0876e7ca3a35", "USD", 1.0},
	{"00cbe9a7-f89f-41b3-a5a6-5bd46d84eaf9", "USD", 1.0},
	{"0384dec7-c147-43cf-bde8-a0fbcf0d2675", "GBP", 1.0},
	{"d6c7f4ec-dab0-4736-a28e-0ab780c7616d", "AUD", 1.0},
	{"09967a27-e3a1-4cf8-81e6-9467126ea2db", "CHF", 1.0},
	{"8666fd78-81e7-4e91-a093-674f57479d06", "EUR", 1.0},
	{"8919a490-bcc8-4f9d-bc42-75af382e6c1b", "USD", 1.0},
	{"2a7574e4-c9f2-4fe3-be8a-95b82faf1056", "USD", 1.0},
	{"75cb0d9e-3d8e-4fc3-be92-f55ae78dd8c4", "EUR", 1.0},

	// Convertible.
	{"7b4be3f7-0de7-4047-b3f2-fec05fdf940c", "GBP", 1.0},
	{"084bbba6-5129-46e3-95fb-4788e300cbc4", "USD", 1.0},
	{"320ebcee-a535-40ea-8634-012691f1404c", "EUR", 1.0},

	// BRCs.
	{"60f92606-29f9-474c-8f1e-91b63669f92b", "USD", 1.0},
	{"4b382cdc-9097-4886-8127-63130e313429", "HKD", 1.0},
	{"661e5163-fbd8-411e-83ee-d5d551fad998", "USD", 1.0},
	{"5959446d-863b-4daf-bd28-0bb87297fbe6", "EUR", 1.0},
	// {"26544020-2b9c-469b-b2cf-d93c1e3c3578", "USD", 1.0}, */
}
