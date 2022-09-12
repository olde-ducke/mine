# Terminal Minesweeper

```console
$ ./mine 
 @  .  @  .  .  .  .  .  .  .  @  .  .  .  .  .  .  .  .  .
 .  .  @  @  .  .  @  .  @  .  .  .  .  @  .  @  @  .  .  .
 @  2  2  2  2  @  .  .  .  .  .  .  .  .  @  .  @  .  .  @
 .  1        2  .  .  .  @  .  @  .  .  .  .  .  .  .  .  .
 .  2  1     1  @  .  %  %  .  .  @  .  .  .  @  .  .  .  @
 @  @  2  1  2  .  .  .  .  @  @  .  .  .  .  .  .  .  @  .
 .  .  .  @  .  .  .  .  @  .  .  @  @  .  .  @  1  1  1  1
 .  @  .  .  .  .  .  .  .  2  1  3  @  2  1  1  1         
 .  .  .  .  .  .  .  @  @  1     1  1  1                  
 .  . [@] .  .  @  .  .restart? [y/n]      1  2  2  1      
 .  .  .  @  .  @  @  .  .  2  1           1  @  @  3  1  1
 .  .  .  .  %  .  .  @  @  @  1  1  1  1  1  .  @  .  @  .
 .  2  @  .  @  .  .  .  .  .  .  .  @  .  .  .  @  .  .  @
 @  .  .  .  2  @  .  .  .  %  .  @  .  .  %  .  .  .  @  .
 @  @  @  .  .  .  @  .  .  .  .  @  .  .  .  .  @  .  .  .
 @  .  @  @  .  1  .  .  .  .  .  %  .  .  .  .  @  .  .  .
 @  .  .  .  .  .  .  @  .  @  .  .  .  .  .  .  .  .  .  .
 .  .  .  @  .  @  .  @  .  .  .  .  @  .  .  .  .  .  .  .
 .  .  .  @  .  .  .  @  .  .  .  .  .  .  .  @  .  .  .  .
 @  .  .  .  .  @  .  .  .  .  @  .  .  @  .  .  .  .  .  .
```

Stolen from [here](https://github.com/tsoding/mine).

Created to better understand how to write interactive terminal application, and play with the idea of building golang code through make.

## Description

Regular minesweeper for terminal, written in go. Nothing special.

|          Representation          | Description   |
|----------------------------------|---------------|
|           `.`                    | closed cell   |
|           ` `                    | empty cell    |
|           `@`                    | bomb          |
|           `%`                    | flag          |


## Build

```console
go build .
```

## Controls

|                     Shortcut                     | Description                                            |
|--------------------------------------------------|--------------------------------------------------------|
|           <kbd>Space</kbd><kbd>f</kbd>           | open field                                             |
|             <kbd>Esc</kbd><kbd>q</kbd>           | quit                                                   |
|              <kbd>Up</kbd><kbd>W</kbd>           | move right                                             |
|            <kbd>Down</kbd><kbd>S</kbd>           | move down                                              |
|            <kbd>Left</kbd><kbd>A</kbd>           | move left                                              |
|           <kbd>Right</kbd><kbd>D</kbd>           | move right                                             |
|                           <kbd>R</kbd>           | restart                                                |
|                           <kbd>Y</kbd>           | confirm                                                |
|                           <kbd>N</kbd>           | decline                                                |
