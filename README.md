# Obsidian Tooling In Go

### About

This is a project that I'm using to learn Go, and to automate the tedious things I don't want to do manually in Obsidian.

It automatically backs up notes to GitHub so you don't lose them, and can clear completed tasks throughout the entire vault using a recursive function.

### How to use

I have a compiled binary named `obsidianautomation` for MacOS named because that's what I use to program, but if you are on Windows or Linux you're gonna have to manually compile the code.

The project is written in Go, so go to https://go.dev/ and install language if you haven't already.

I'm not too good with how the Go dependency management works, but probably run `go mod tidy` and then `go build`

Once it's built, place the binary in the root of an Obsidian vault (really you could use this program with any markdown note taking system, but I use Obsidian!!!) and run the binary.

### I love Obsidian!!!

I would encourage anyone that uses Obsidian to look into this code! I think everyone can benefit from the functionality it has even at this primitive stage.
