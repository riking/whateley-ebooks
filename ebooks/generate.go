package ebooks

//go:generate go-bindata -nomemcopy story.css
//go:generate sed -i "s/package main/package ebooks/" bindata.go
