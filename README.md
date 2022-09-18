# FontXplorer
A tool to preview old DOS FONTX fonts for Japanese (SHIFT-JIS)

## Purpose
It has become necessary to support Japanese language in my projects and to display Japanese text (kanji and kana) on graphics LCDs.
To this end, I was advised by a Japanese developer (Elm-chan) to do what they do in Japan - i.e. use old FONTX files from DOS 5.x.
In Japan the wheel is not reinvented, nor do they incur the overhead of UNICODE in an embedded project. Instead, they rely on good old SHIFT-JIS.
The FONTX files are used directly in embedded projects, either being read from an SD card or FLASH.

## Usage
Simply run the program with a valid FONTX filename as an argument. Various information about the font will be printed, and then the character glyphs
will be displayed

## Credits
http://elm-chan.org/docs/dosv/fontx_e.html
助けてくれたElm-ちゃんに感謝します。 
GRAPHLCDで日本語を扱えるようにするというこの問題を解決する方法を私は何年も考えてきました
