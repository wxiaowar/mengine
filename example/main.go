package main

import "mengine"


// /i/module/func
func CheckIntegrity (ctx *mengine.Context) (e error) {
	return
}

func main() {
	opt := &mengine.EngionOption{
		IsDebug: true,
		Addr: "127.0.0.1:8080",
		CheckWhiteList: nil,
		CheckIntegrity: CheckIntegrity,
	}

	r := &router{

	}

	log := &mlog{

	}

	eg := mengine.NewEngine(opt, r, log)
	eg.Run()
}