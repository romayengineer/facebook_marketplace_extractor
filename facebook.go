package main

type FacebookScrapperInterface interface {
	Login(username string, password string)
}

type FacebookScrapper struct {
}

func (fs *FacebookScrapper) Login(username string, passwrod string) {

}
