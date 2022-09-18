package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

// Program variables

// Font Information
// Refer to http://elm-chan.org/docs/dosv/fontx_e.html
var signature = make([]byte, 6)
var fontname = make([]byte, 8)
var width byte
var height byte
var code_flag byte
var code_blocks byte

type block struct {
	blockstart uint16
	blockend   uint16
}
type codeblocktable struct {
	blockentry [255]block
}

var codetable codeblocktable

//----------------------------------------------------------------------------------
// Functions
//----------------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// Name: byte2glyph
// Function: Render the byte, horizontally as a glyph using asterisks and dashes
// Parameters: byte to glyph
// Returns: formatted string
//-----------------------------------------------------------------------------
func byte2glyph(data byte) string {
	var s string = ""
	for p := 0; p < 8; p++ {
		if data&0x80 == 0x80 {
			s += "█"
		} else {
			s += " "
		}
		data = data << 1
	}
	return s
}

// Name: ReadFont
// Function: Read entire font file to memory
// Parameter: filename (string type)
// Returns: Byte array holding file contents, error
//-------------------------------------------------------
func ReadFont(filename string) ([]byte, error) {
	file, err := os.Open(filename)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return nil, statsErr
	}
	var size int64 = stats.Size()
	rawbytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(rawbytes)

	return rawbytes, err
}

//-----------------------------------------------------------------------------------
// Main function
//------------------------------------------------------------------------------------
func main() {

	// Store program argument(s) the main one is the FONTX file name
	// If no argument, variable 'args' will have no length but will NOT be = nil!
	args := os.Args[1:]

	fmt.Println(" ")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("|                             FontXplorer                              |")
	fmt.Println("|                   Utility to preview DOS FontX files                 |")
	fmt.Println("|                              By Sonic2k                              |")
	fmt.Println("+----------------------------------------------------------------------+")
	fmt.Println("  ")

	if len(args) == 0 {
		fmt.Println(" [ERROR] no filename given... exiting")
		return // Terminate program
	}
	// Filename is given, let's try and open the font
	var filename string
	filename = args[0]
	fontdata, err := ReadFont(filename)
	if err != nil {
		fmt.Println(" [ERROR] unable to open file... exiting")
		return
	}
	// Font file loaded OK, let's begin to parse
	var fontpointer int = 0

	// Parse out and check signature
	for n := 0; n < 6; n++ {
		signature[n] = fontdata[fontpointer]
		fontpointer++
	}
	var s string = string(signature[:])
	if s == "FONTX2" {
		// We have a valid signature, now we can start parsing things out...

		// Parse out the font name
		for n := 0; n < 8; n++ {
			fontname[n] = fontdata[fontpointer]
			fontpointer++
		}

		// Parse out width and height, code flag
		width = fontdata[fontpointer]
		fontpointer++
		height = fontdata[fontpointer]
		fontpointer++
		code_flag = fontdata[fontpointer]
		fontpointer++
		code_blocks = fontdata[fontpointer]
		fontpointer++

		// Fill in the code block table
		for n := 0; n < int(code_blocks); n++ {
			data := []byte{fontdata[fontpointer], fontdata[fontpointer+1]}
			fontpointer += 2
			codetable.blockentry[n].blockstart = binary.LittleEndian.Uint16(data)
			data = []byte{fontdata[fontpointer], fontdata[fontpointer+1]}
			fontpointer += 2
			codetable.blockentry[n].blockend = binary.LittleEndian.Uint16(data)
		}

		fmt.Println("                Font Information                  ")
		fmt.Println("    ----------------------------------------------")
		fmt.Println("         Font Name: " + string(fontname[:]))
		fmt.Println(fmt.Sprintf("         Character Width:  %d pixels", width))
		fmt.Println(fmt.Sprintf("         Character Height: %d pixels", height))
		fmt.Println(fmt.Sprintf("         Code flag: %d ", code_flag))
		if code_flag == 1 {
			fmt.Println("                    └ Shift-JIS")
		} else {
			fmt.Println("                    └ ANSI")
		}
		fmt.Println(fmt.Sprintf("         Number of code blocks: %d ", code_blocks))
		for n := 0; n < int(code_blocks); n++ {
			fmt.Println(fmt.Sprintf("                Block #%2d :    Block Start: %04X    Block End: %04X   ", n+1,
				codetable.blockentry[n].blockstart, codetable.blockentry[n].blockend))
		}

		// Begin glyph dumping
		fmt.Println("       ")
		fmt.Println("                Font Glyphs                                ")
		fmt.Println("    -------------------------------------------------------")
		fmt.Println(fmt.Sprintf("      Current pointer location in file: %d   0x%04X", fontpointer, fontpointer))

		// This is test code... we know the font is 16 x 16 so to get a quick overview we will just dump using 16 x 16
		for r := 0; r < 1524; r++ { // outer loop... 16 chars
			fmt.Println(fmt.Sprintf(" Char Offset: %d", r))
			for q := 0; q < 16; q++ { // inner loop... 16 cols of 2 bytes
				fmt.Println(fmt.Sprintf(" %s%s", byte2glyph(fontdata[fontpointer]), byte2glyph(fontdata[fontpointer+1])))
				fontpointer += 2
			}

			fmt.Println("         ")
		}

	} else {
		fmt.Println(" [ERROR] Invalid font signature. Is this a valid FONTX2 font file?   Exiting")
		return
	}

}
