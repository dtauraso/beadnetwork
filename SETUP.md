# Setup from a brand-new Mac

[← README.md](README.md)

If you already have Go, Node, VS Code and a clone of the repo, skip to
[Build in README.md](README.md#build). Otherwise start here — no prior setup assumed.

Everything below is typed into **Terminal**. To open it: press `Cmd`+`Space`,
type `Terminal`, press `Return`. You get a window where you type a command and
press `Return` to run it. Commands are shown one per line; run them in order.

Some commands ask for your Mac login password. Nothing appears on screen as you
type it — that is normal, not a frozen terminal. Type it and press `Return`.

## 1. Apple's developer tools

This gives you `git`, which downloads the project.

```sh
xcode-select --install
```

A dialog appears — click **Install** and wait for it to finish (a few minutes).
If it instead says `command line tools are already installed`, you already have
them; move on.

## 2. Homebrew

Homebrew installs the other three tools. Paste this whole line:

```sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

When it finishes it prints a short **Next steps** section telling you to run two
more commands (an `echo` and an `eval`). **Run them.** On Apple Silicon Macs,
Homebrew is not usable until you do. Then confirm it worked:

```sh
brew --version
```

If that prints a version number, continue. If it says `command not found`, the
Next steps commands were not run — scroll up in the terminal and run them.

## 3. Go, Node, and VS Code

```sh
brew install go
brew install node
brew install --cask visual-studio-code
```

Go is needed both to build and to *run* the editor: the extension shells out to
the Go compiler every time you open it.

Now open VS Code once (`Cmd`+`Space`, type `Visual Studio Code`, `Return`). Inside
it press `Cmd`+`Shift`+`P`, type `shell command`, and choose
**Shell Command: Install 'code' command in PATH**. This makes `code` work in
Terminal.

## 4. Download the project

```sh
cd ~/Documents
git clone https://github.com/dtauraso/beadnetwork.git
cd beadnetwork
```

You are now inside the project folder. Later terminal sessions start back at your
home folder, so run `cd ~/Documents/beadnetwork` again to return here.

## 5. Check everything landed

```sh
go version
node -v
code -v
```

Three version numbers means you are ready. Any `command not found` means that
tool's step above did not complete.

Continue with [Build in README.md](README.md#build).
