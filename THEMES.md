# Cozy Themes Guide

Cozy supports customizable themes for a personalized reading experience.

## Built-in Themes

### 1. **cozy-dark** (Default)
A warm, purple-tinted dark theme perfect for night reading.
- Primary: Soft purple (#A78BFA)
- Background: Dark blue-gray (#1F2937)
- Designed for comfort during extended reading sessions

### 2. **solarized-dark**
The classic Solarized Dark color scheme, scientifically designed for readability.
- Primary: Blue (#268BD2)
- Background: Deep blue-green (#002B36)
- Popular among developers and readers

### 3. **sepia**
A warm, book-like theme reminiscent of old paper.
- Primary: Saddle brown (#8B4513)
- Background: Sepia (#F5E6D3)
- Perfect for daytime reading with reduced eye strain

## Using Themes

### Switch to a Built-in Theme

Edit your config file at `~/.config/cozy/config.toml`:

```toml
theme_name = "solarized-dark"
```

Available options:
- `"cozy-dark"`
- `"solarized-dark"`
- `"sepia"`

## Creating Custom Themes

### Step 1: Create a Theme File

Create a new file in `~/.config/cozy/themes/` with the `.toml` extension.

For example: `~/.config/cozy/themes/my-theme.toml`

### Step 2: Define Your Theme

```toml
name = "my-theme"

# UI Colors
primary_color = "#FF6B6B"      # Main accent color (headers, highlights)
secondary_color = "#4ECDC4"    # Secondary accent (progress, metadata)
background_color = "#1A1A2E"   # Main background

# Text Colors
text_color = "#EAEAEA"         # Primary text color
muted_text_color = "#95A5A6"   # Secondary/dimmed text

# Element-Specific Colors
heading_color = "#FFE66D"      # Chapter titles, headings (h1-h6)
link_color = "#48B9FF"         # Hyperlinks
quote_color = "#BDC3C7"        # Blockquote text
quote_border_color = "#E74C3C" # Blockquote left border
code_bg_color = "#2C3E50"      # Code block background
code_text_color = "#2ECC71"    # Code text color
emphasis_color = "#F39C12"     # Italic/emphasized text
strong_color = "#E67E22"       # Bold/strong text
```

### Step 3: Use Your Custom Theme

Update your config file:

```toml
theme_name = "my-theme"
```

Restart Cozy to see your new theme in action!

## Theme Color Reference

### What Each Color Controls

| Color | Usage | Example |
|-------|-------|---------|
| `primary_color` | Book titles, main UI accents | Top header, selected items |
| `secondary_color` | Chapter info, progress indicators | "Chapter 3/15", scroll percentage |
| `background_color` | Main background | (Terminal background override) |
| `text_color` | Body text | Paragraphs, main content |
| `muted_text_color` | Help text, subtle UI elements | Horizontal rules, metadata |
| `heading_color` | Chapter titles, h1-h6 tags | "# Chapter One" |
| `link_color` | Hyperlinks in EPUB | Links, references |
| `quote_color` | Blockquote text | `<blockquote>` content |
| `quote_border_color` | Blockquote left border | Visual indicator for quotes |
| `code_bg_color` | Code block background | `<pre>`, `<code>` blocks |
| `code_text_color` | Code text | Code content |
| `emphasis_color` | Emphasized/italic text | `<em>`, `<i>` tags |
| `strong_color` | Bold/strong text | `<strong>`, `<b>` tags |

## Rich Formatting Support

Cozy renders HTML elements with custom styling:

### Headings
- **H1**: Bold, underlined, large
- **H2-H6**: Bold, progressively smaller
- All headings use `heading_color`

### Text Formatting
- **Bold** (`<strong>`, `<b>`): Uses `strong_color`, bold weight
- *Italic* (`<em>`, `<i>`): Uses `emphasis_color`, italic style
- `Code` (`<code>`): Uses `code_text_color` on `code_bg_color` background

### Block Elements
- **Blockquotes** (`<blockquote>`):
  - Italic text in `quote_color`
  - Left border in `quote_border_color`
  - Indented with padding

- **Code Blocks** (`<pre>`):
  - Monospace font
  - Background in `code_bg_color`
  - Text in `code_text_color`
  - Padded for readability

- **Horizontal Rules** (`<hr>`):
  - Line of dashes
  - Uses `muted_text_color`

### Lists
- Bullet points for `<ul>`
- Proper indentation for nested lists
- Uses primary `text_color`

## Tips for Creating Themes

### Color Contrast
- Ensure sufficient contrast between text and background (4.5:1 ratio minimum)
- Test your theme with different content types

### Color Harmony
- Use a color palette generator (coolors.co, paletton.com)
- Stick to 3-5 main colors with variations

### Readability
- Body text should be comfortable for extended reading
- Headings should stand out but not overpower
- Code blocks should be easily distinguishable

### Inspiration
- Terminal color schemes (Dracula, Nord, Gruvbox)
- E-reader apps (Kindle, Apple Books themes)
- Popular IDE themes (VS Code, Sublime Text)

## Example Themes

### Nord-inspired
```toml
name = "nord"
primary_color = "#88C0D0"
secondary_color = "#81A1C1"
background_color = "#2E3440"
text_color = "#ECEFF4"
muted_text_color = "#4C566A"
heading_color = "#8FBCBB"
link_color = "#5E81AC"
quote_color = "#D8DEE9"
quote_border_color = "#88C0D0"
code_bg_color = "#3B4252"
code_text_color = "#A3BE8C"
emphasis_color = "#EBCB8B"
strong_color = "#BF616A"
```

### Dracula-inspired
```toml
name = "dracula"
primary_color = "#BD93F9"
secondary_color = "#FF79C6"
background_color = "#282A36"
text_color = "#F8F8F2"
muted_text_color = "#6272A4"
heading_color = "#8BE9FD"
link_color = "#8BE9FD"
quote_color = "#F8F8F2"
quote_border_color = "#FF79C6"
code_bg_color = "#44475A"
code_text_color = "#50FA7B"
emphasis_color = "#FFB86C"
strong_color = "#FF5555"
```

## Troubleshooting

### Theme Not Loading
1. Check the file is in `~/.config/cozy/themes/`
2. Verify the filename matches `theme_name` in config (without .toml)
3. Check TOML syntax is valid
4. Cozy will fall back to `cozy-dark` if theme fails to load

### Colors Look Wrong
- Some terminals don't support true color (24-bit color)
- Try a different terminal emulator (iTerm2, Alacritty, WezTerm)
- Check your `$TERM` environment variable

### Want to Reset?
Delete or rename your custom theme file and set:
```toml
theme_name = "cozy-dark"
```

## Sharing Your Theme

Created an awesome theme? Share it with the community!
- Post in GitHub Discussions
- Create a PR to add it as a built-in theme
- Share on social media with `#CozyReader`
