package chess_engine

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

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

func createImage(width int, height int, dark, light color.RGBA) *image.RGBA {
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	draw.Draw(img, img.Bounds(), &image.Uniform{dark}, image.ZP, draw.Src)
	return img
}

// Size should be a multiple of 8
func createBoard(size int, dark, light color.RGBA) *image.RGBA {
	img := createImage(size, size, dark, light)
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
			x := (i % 8)
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
	// TODO embed file
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

var Chess_Font = loadTTF()

func BoardToImage(board Board) *image.Paletted {
	img := createBoard(256, color.RGBA{0x44, 0x44, 0xaa, 0xff}, color.RGBA{0xaa, 0xaa, 0xaa, 0xff})
	drawPieces(256, img, Chess_Font, board)
	palettedImage := image.NewPaletted(img.Bounds(), palette.Plan9)
	draw.Draw(palettedImage, palettedImage.Rect, img, img.Bounds().Min, draw.Over)
	return palettedImage
}

func BoardToPNG(board Board, file string) error {
	img := BoardToImage(board)
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	return png.Encode(out, img)
}

func BoardToGIF(board Board, file string) error {
	img := BoardToImage(board)
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	opts := &gif.Options{NumColors: 255}
	return gif.Encode(out, img, opts)
}

func MovesToGIF(startingPosition *Game, moves []*Move, file string, delay int) error {
	images := []*image.Paletted{}
	delays := []int{}
	game := startingPosition
	for _, move := range moves {
		images = append(images, BoardToImage(game.Board))
		delays = append(delays, delay)
		game = game.ApplyMove(move)
	}
	images = append(images, BoardToImage(game.Board))
	delays = append(delays, delay)

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	opts := &gif.GIF{
		Image:           images,
		Delay:           delays,
		LoopCount:       0,
		BackgroundIndex: 0,
	}
	if err := gif.EncodeAll(out, opts); err != nil {
		return err
	}
	out.Close()
	return exec.Command("convert", file, "-coalesce", file).Run()
}

func init() {

	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	game, err := ParseFEN(fen)
	if err != nil {
		panic(err)
	}
	BoardToPNG(game.Board, "test.png")
	if err := BoardToGIF(game.Board, "test.gif"); err != nil {
		panic(err)
	}
	if err := MovesToGIF(game, []*Move{NewMove(E2, E4), NewMove(E7, E5)}, "test2.gif", 100); err != nil {
		panic(err)
	}
}
