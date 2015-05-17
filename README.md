# ded
Simple text editor that works with 9P server.

# Why?

* Curiosity and pratice!.
* Scratching my own itches.
* Using Go to create graphical user interfaces with [gxui](https://github.com/google/gxui)

Sometimes it's rellay hard for programs interact with your text editor.
Ded solves that problem by exposing all editor internals as "files".
Those files can be manipulated by any 9P client or programs that interact with stdin/stdout.

If you think this looks like [Acme](http://acme.cat-v.org/), then your are right. Acme was the inspiration for Ded interaction.
Check this [video from Russ Cox](http://research.swtch.com/acme) about Acme.

For instance: How to sort all lines in my current window?

Consider that "./cat" refeers to the cat command shipped with Ded.

`./cat /body | sort | ./cat -w /body`

Run that in any shell and you have sorted all the contents of the editor.

Another example: How to save or load files

`cat "your file" | ./cat -w /body` -> this loads "your file" from disk and displays it on the editor window.

`./cat /body > "your file"` -> save the changes to "your file"

# What is missing?

A lot of stuff, for example having more than one window open. I am focusing on improving the interaction with 9P clients, before starting to worry about the graphical user interaction.

# What Ded means

Dayana's Editor [dayanafonseca.com](http://www.dayanafonseca.com/).
