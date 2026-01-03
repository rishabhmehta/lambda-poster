package poster

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	"golang.org/x/image/font"
)

//go:embed assets/picture.jpg
//go:embed assets/poppins.ttf
var Assets embed.FS

const (
	avatarSize = 150 // Avatar width/height in pixels
	fontSize   = 36  // Font size for name
	dpi        = 72  // DPI for font rendering
)

// Generator handles poster image generation
type Generator struct {
	background image.Image
	font       *truetype.Font
}

// NewGenerator creates a new poster generator with embedded assets
func NewGenerator() (*Generator, error) {
	// Load background image
	bgData, err := Assets.ReadFile("assets/picture.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to read background: %w", err)
	}

	bg, err := jpeg.Decode(bytes.NewReader(bgData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode background: %w", err)
	}

	// Load font
	fontData, err := Assets.ReadFile("assets/poppins.ttf")
	if err != nil {
		return nil, fmt.Errorf("failed to read font: %w", err)
	}

	f, err := truetype.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	return &Generator{
		background: bg,
		font:       f,
	}, nil
}

// Generate creates a poster with the given name and avatar
func (g *Generator) Generate(name, avatarURL string) (string, error) {
	// Fetch avatar
	avatar, err := fetchImage(avatarURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}

	// Resize avatar to fixed size
	avatar = resize.Resize(avatarSize, avatarSize, avatar, resize.Lanczos3)

	// Create output image
	bounds := g.background.Bounds()
	output := image.NewRGBA(bounds)

	// Draw background
	draw.Draw(output, bounds, g.background, image.Point{}, draw.Src)

	// Calculate avatar position (center horizontally, upper third vertically)
	avatarBounds := avatar.Bounds()
	avatarX := (bounds.Dx() - avatarBounds.Dx()) / 2
	avatarY := bounds.Dy()/3 - avatarBounds.Dy()/2

	// Draw avatar
	avatarRect := image.Rect(avatarX, avatarY, avatarX+avatarBounds.Dx(), avatarY+avatarBounds.Dy())
	draw.Draw(output, avatarRect, avatar, image.Point{}, draw.Over)

	// Draw name below avatar
	if err := g.drawText(output, name, avatarY+avatarBounds.Dy()+30); err != nil {
		return "", fmt.Errorf("failed to draw text: %w", err)
	}

	// Encode to PNG and base64
	var buf bytes.Buffer
	if err := png.Encode(&buf, output); err != nil {
		return "", fmt.Errorf("failed to encode output: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// drawText draws centered text at the given Y position
func (g *Generator) drawText(img *image.RGBA, text string, y int) error {
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(g.font)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)
	c.SetHinting(font.HintingFull)

	// Measure text width for centering
	opts := truetype.Options{Size: fontSize, DPI: dpi}
	face := truetype.NewFace(g.font, &opts)
	textWidth := font.MeasureString(face, text).Ceil()

	// Calculate X position for center alignment
	x := (img.Bounds().Dx() - textWidth) / 2

	pt := freetype.Pt(x, y)
	_, err := c.DrawString(text, pt)
	return err
}

// fetchImage downloads and decodes an image from URL
func fetchImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}
