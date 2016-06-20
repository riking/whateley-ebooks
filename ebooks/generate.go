package ebooks

//go:generate go-bindata -nomemcopy -debug content.opf cover.html part.html story.css toc.ncx
//go:generate sed -i "s/package main/package ebooks/" bindata.go
