from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter


ROOT = Path(__file__).resolve().parents[1]
SIZE = 1024


def vertical_gradient(top, bottom):
    image = Image.new("RGBA", (SIZE, SIZE))
    pixels = image.load()
    for y in range(SIZE):
        ratio = y / (SIZE - 1)
        color = tuple(round(top[i] * (1 - ratio) + bottom[i] * ratio) for i in range(4))
        for x in range(SIZE):
            pixels[x, y] = color
    return image


def build_icon():
    canvas = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    shadow = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    shadow_draw.rounded_rectangle((70, 72, 954, 956), radius=224, fill=(33, 104, 223, 52))
    canvas.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(32)))

    mask = Image.new("L", canvas.size, 0)
    mask_draw = ImageDraw.Draw(mask)
    mask_draw.rounded_rectangle((64, 52, 960, 948), radius=224, fill=255)
    surface = vertical_gradient((255, 255, 255, 255), (237, 245, 255, 255))
    canvas.alpha_composite(Image.composite(surface, Image.new("RGBA", canvas.size), mask))

    draw = ImageDraw.Draw(canvas)
    draw.rounded_rectangle((64, 52, 960, 948), radius=224, outline=(186, 213, 255, 255), width=26)
    draw.rounded_rectangle((88, 76, 936, 924), radius=202, outline=(255, 255, 255, 230), width=12)

    draw.line([(288, 746), (500, 282), (724, 746)], fill=(47, 124, 246, 255), width=92, joint="curve")
    for point in ((288, 746), (500, 282), (724, 746)):
        draw.ellipse((point[0] - 46, point[1] - 46, point[0] + 46, point[1] + 46), fill=(47, 124, 246, 255))

    signal = [(324, 622), (432, 622), (506, 516), (582, 598), (724, 428)]
    draw.line(signal, fill=(34, 195, 238, 255), width=54, joint="curve")
    for point in signal:
        draw.ellipse((point[0] - 27, point[1] - 27, point[0] + 27, point[1] + 27), fill=(34, 195, 238, 255))
    draw.line([(680, 427), (746, 407), (732, 475)], fill=(47, 124, 246, 255), width=42, joint="curve")
    draw.ellipse((472, 254, 528, 310), fill=(139, 220, 255, 255), outline=(255, 255, 255, 255), width=12)
    return canvas


def main():
    icon = build_icon()
    (ROOT / "build" / "darwin").mkdir(parents=True, exist_ok=True)
    icon.save(ROOT / "build" / "appicon.png", optimize=True)
    icon.save(
        ROOT / "build" / "app.ico",
        format="ICO",
        sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)],
    )
    icon.save(
        ROOT / "build" / "windows" / "icon.ico",
        format="ICO",
        sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)],
    )
    icon.save(ROOT / "build" / "darwin" / "icon.icns", format="ICNS")


if __name__ == "__main__":
    main()
