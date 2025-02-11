package controllers

type ResponseSuccess struct {
	Status int    `example:"201"`
	Title  string `example:"Record successfully created"`
	Type   string `example:"/status/success"`
}

type ResponseSuccessCreated struct {
	Status int    `example:"201"`
	Title  string `example:"Record successfully created"`
	Type   string `example:"/status/success"`
}

type ResponseSuccessDeleted struct {
	Status int    `example:"200"`
	Title  string `example:"Record successfully deleted"`
	Type   string `example:"/status/success"`
}

type ResponseSuccessUpdated struct {
	Status int    `example:"200"`
	Title  string `example:"Record successfully updated"`
	Type   string `example:"/status/success"`
}

type ResponseError struct {
	Type   string `example:"/errors/schema-validation"`
	Title  string `example:"Schema validation failed"`
	Status int    `example:"400"`
	Error  string `example:"Missing required fields"`
}
