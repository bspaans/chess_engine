package chess_engine

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

func readPiecesImage() image.Image {

	reader, err := os.Open("pieces.png")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	return m
}

func createImage(width int, height int, background color.RGBA) *image.RGBA {
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), &image.Uniform{background}, image.ZP, draw.Src)
	return img
}

// Size should be a multiple of 8
func createBoard(size int, dark, light color.RGBA) *image.RGBA {
	img := createImage(size, size, dark)
	squareWidth := size / 8
	for x := 0; x < 8; x += 2 {
		for y := 0; y < 8; y += 2 {
			bounds := image.Rect(x*squareWidth, y*squareWidth, x*squareWidth+squareWidth, y*squareWidth+squareWidth)
			c := light
			draw.Draw(img, bounds, &image.Uniform{c}, image.ZP, draw.Src)
		}
	}
	for x := 1; x < 8; x += 2 {
		for y := 1; y < 8; y += 2 {
			bounds := image.Rect(x*squareWidth, y*squareWidth, x*squareWidth+squareWidth, y*squareWidth+squareWidth)
			c := light
			draw.Draw(img, bounds, &image.Uniform{c}, image.ZP, draw.Src)
		}
	}
	return img
}

// TODO: only works properly at size 256. Font scaling be hard.
func drawPieces(size int, img *image.RGBA, piecesFont *truetype.Font, board Board) {
	squareWidth := size / 8
	for i := 0; i < 64; i++ {
		if board[i] != NoPiece {
			x := 7 - (i % 8)
			y := 8 - (i / 8)

			pieces := map[Piece]string{
				WhiteKing:   "l",
				BlackKing:   "l",
				WhiteKnight: "j",
				BlackKnight: "j",
				WhiteBishop: "n",
				BlackBishop: "n",
				WhiteQueen:  "w",
				BlackQueen:  "w",
				WhitePawn:   "o",
				BlackPawn:   "o",
				WhiteRook:   "t",
				BlackRook:   "t",
			}

			ctx := freetype.NewContext()
			ctx.SetFont(piecesFont)
			ctx.SetDPI(float64(size))
			ctx.SetFontSize(8)
			ctx.SetDst(img)
			ctx.SetClip(img.Bounds())

			if board[i].Color() == Black {
				ctx.SetSrc(image.Black)
			} else {
				ctx.SetSrc(image.White)
			}
			ctx.DrawString(pieces[board[i]], freetype.Pt(x*squareWidth+2, y*squareWidth-5))

		}
	}
}

func loadTTF() *truetype.Font {
	ttf, err := ioutil.ReadFile("chess.ttf")
	if err != nil {
		panic(err)
	}
	font, err := freetype.ParseFont(ttf)
	if err != nil {
		panic(err)
	}
	return font
}

func init() {
	ttf := loadTTF()
	img := createBoard(256, color.RGBA{0x66, 0x66, 0x66, 0xff}, color.RGBA{0xaa, 0xaa, 0xaa, 0xff})

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	game, err := ParseFEN(fen)
	if err != nil {
		panic(err)
	}
	drawPieces(256, img, ttf, game.Board)

	out, err := os.Create("test.png")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	png.Encode(out, img)
}
