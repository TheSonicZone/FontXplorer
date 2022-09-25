package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

// Program variables
//------------------------------------------------------------------------
// Font Information
// Refer to http://elm-chan.org/docs/dosv/fontx_e.html
var signature = make([]byte, 6)
var fontname = make([]byte, 8)
var width uint16
var height uint16
var code_flag byte
var code_blocks byte

type block struct {
	blockstart uint16
	blockend   uint16
	char_range uint16
}
type codeblocktable struct {
	blockentry [255]block
}

var codetable codeblocktable

//----------------------------------------------------------------------------------
// Functions
//----------------------------------------------------------------------------------

//----------------------------------------------------------------------------------
// Name: ExpandToBoundary
// Function: Expands a given font's width to a byte boundary
// Parameters: Width specified in font file
// Returns: Expanded width that will fit into byte boundaries
//-----------------------------------------------------------------------------------
func ExpandToBoundary(width uint16) uint16 {
	x := width
	y := x % 8 // this is useful for determining overlap
	z := x / 8
	if y != 0 {
		z++
	}
	return z * 8
}

//-----------------------------------------------------------------------------
// Name: byte2glyph
// Function: Render the byte, horizontally as a glyph using codepage 437 characters
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

//-------------------------------------------------------
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

//-------------------------------------------------------------------------
// Name: lookup_block
// Function: Find the block in which the given Shift-JIS code falls within
// Parameters: Shift-JIS code, Pointer to location where offset shall be written
// Returns: Number of the code table in which the Shift-JIS code resides
//          returns -1 if given Shift-JIS code is not located in any of the code tables
//--------------------------------------------------------------------------
func lookup_block(jis uint16, offset *uint32) int {

	var low_end uint16
	var high_end uint16
	var code_range uint16
	var code_total uint32 = 0
	for n := 0; n < int(code_blocks); n++ {
		low_end = codetable.blockentry[n].blockstart
		high_end = codetable.blockentry[n].blockend
		if n > 0 {
			code_range = codetable.blockentry[n-1].char_range
			code_total += uint32(code_range)
		}
		if jis >= low_end && jis <= high_end {
			*offset = code_total
			return n
		}

	}
	return -1
}

//----------------------------------------------------------------------------------
// Name: showChar
// Function: Render the character glyph to the console
// Parameters: Address in file where data for this particular glyph begins
//             Pointer to array holding file data
//             Width and Height
// Returns: nil
//----------------------------------------------------------------------------------
func showChar(address uint32, src []byte, width uint16, height uint16) {

	var address_in_file = address
	for q := 0; q < int(width); q++ {
		if height == 32 {
			fmt.Println(fmt.Sprintf(" %s%s%s%s",
				byte2glyph(src[address_in_file]),
				byte2glyph(src[address_in_file+1]),
				byte2glyph(src[address_in_file+2]),
				byte2glyph(src[address_in_file+3])))
		}
		if height == 24 {
			fmt.Println(fmt.Sprintf(" %s%s%s", byte2glyph(src[address_in_file]), byte2glyph(src[address_in_file+1]), byte2glyph(src[address_in_file+2])))
			address_in_file += 3
		}
		if height == 16 {
			fmt.Println(fmt.Sprintf(" %s%s", byte2glyph(src[address_in_file]), byte2glyph(src[address_in_file+1])))
			address_in_file += 2
		}
		if height == 8 {
			fmt.Println(fmt.Sprintf(" %s", byte2glyph(src[address_in_file])))
			address_in_file += 1
		}

	}

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
	fmt.Println("|                 Utility to preview CJK DOS FontX characters          |")
	fmt.Println("|  For development of GraphLCD2 to support Unicode and Asian languages |")
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
		width = uint16(fontdata[fontpointer])
		fontpointer++
		height = uint16(fontdata[fontpointer])
		fontpointer++
		code_flag = fontdata[fontpointer]
		fontpointer++
		code_blocks = fontdata[fontpointer]
		fontpointer++

		// Expand width value to fit into 8 bit rows
		width = ExpandToBoundary(width)

		// Fill in the code block table
		for n := 0; n < int(code_blocks); n++ {
			data := []byte{fontdata[fontpointer], fontdata[fontpointer+1]}
			fontpointer += 2
			codetable.blockentry[n].blockstart = binary.LittleEndian.Uint16(data)
			data = []byte{fontdata[fontpointer], fontdata[fontpointer+1]}
			fontpointer += 2
			codetable.blockentry[n].blockend = binary.LittleEndian.Uint16(data)
			codetable.blockentry[n].char_range = codetable.blockentry[n].blockend - codetable.blockentry[n].blockstart
		}

		fmt.Println("                Font Information                  ")
		fmt.Println("    ----------------------------------------------")
		fmt.Println("         Font Name: " + string(fontname[:]))
		fmt.Println(fmt.Sprintf("         Character Width:  %d pixels", width))
		fmt.Println(fmt.Sprintf("         Character Height: %d pixels", height))
		fmt.Println(fmt.Sprintf("         Code flag: %d ", code_flag))
		if code_flag == 1 {
			fmt.Println("                    ^ Shift-JIS font")
		} else {
			fmt.Println("                    ^ ANK font")
		}

		fmt.Println(fmt.Sprintf("         Number of code blocks: %d ", code_blocks))
		for n := 0; n < int(code_blocks); n++ {
			fmt.Println(fmt.Sprintf("            Block #%2d :    Block Start: %04X    Block End: %04X   Range: %04X", n+1,
				codetable.blockentry[n].blockstart, codetable.blockentry[n].blockend, codetable.blockentry[n].char_range))
		}

		// Begin glyph dumping
		fmt.Println("       ")
		fmt.Println("                Font Glyph Location Calcs                  ")
		fmt.Println("    -------------------------------------------------------")
		fmt.Println(fmt.Sprintf("      Font glyph data start location in file:  0x%04X   (%d)", fontpointer, fontpointer))

		var shift_jis_code uint16 = 0x82A0 // Test value to check lookup performance etc.. Shift-JIS 82A0 = あ
		fmt.Println(fmt.Sprintf("          Shift-JIS code as input: %04X", shift_jis_code))
		var offset uint32
		block_location := lookup_block(shift_jis_code, &offset)
		if block_location == -1 {
			fmt.Println("             Given Shift-JIS code not found in code tables")
		} else {

			offset += uint32(block_location)
			char_start := codetable.blockentry[block_location].blockstart
			char_offset := (shift_jis_code - char_start) * uint16((width/8)*height)
			char_address_base := uint32(fontpointer) + uint32((width/8)*height)*uint32(offset)
			char_effective_address := char_address_base + uint32(char_offset)
			fmt.Println(fmt.Sprintf("          Shift-JIS code located in code table %d    Table Offset: %d (%04X)", block_location+1, offset, offset))
			fmt.Println(fmt.Sprintf("          Shift-JIS first char code in block: %04X     Offset: %d bytes", char_start, char_offset))
			fmt.Println(fmt.Sprintf("          Base address of char range in file: 0x%04X", char_address_base))
			fmt.Println(fmt.Sprintf("          Absolute address of char glyph in file: 0x%04X", char_effective_address))
			fmt.Println("   ")
			fmt.Println(fmt.Sprintf("---------- Char Code (Shift-JIS):%04X --------", shift_jis_code))
			showChar(char_effective_address, fontdata, width, height)
			fmt.Println("--------------------------------------------")
			//return
		}
		fmt.Println("                       ")

	} else {
		fmt.Println(" [ERROR] Invalid font signature. Is this a valid FONTX2 font file?   Exiting")
		return
	}

}
