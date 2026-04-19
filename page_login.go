package main

type PageLogin struct {
	InputUsername string
	InputPassword string
}

var pageLogin = PageLogin{
	InputUsername: "input[name=email]",
	InputPassword: "input[name=pass]",
}
