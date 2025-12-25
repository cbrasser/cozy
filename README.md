# COZY - TUI Ebook reader

 🚨 Work in Progress 🚨

Cozy is (or will be) a simple, keyboard-based terminal application for reading e-books. 

The main goals for `cozy` are:

- Simplicity: Keyboard-based navigation, intuitive shortcuts, 
- Speed: Open up your books without the typical electron-app delay
- Configuration: Use a sensible default color scheme or adapt everything to your needs: colors, font, spacing, ...

## Current State of development

`cozy` currently is:

- Technically already an e-book reader. It can display `e.pub` files in a per-chapter basis with some simple styling. It will generate a config file in `/home/user/.config/cozy/` and user `/home/user/Documents/books` as the default directory to look for files.

## Next steps

- [ ] TOC View -> Display table of contents of current book and navigate to chapters
- [ ] Store reading status -> Remember the last position when re-opening the app
- [ ] Bookmarks -> Save positions in a book with an optional comment / tag
- [ ] Search -> self explanatory
- [ ] Improve Rendering & Styling options
- [ ] Image support -> Maybe generate ASCII from image, will have to do some research on what is possible with different image backends
- [ ] Themes -> Improve theme support
- [ ] More config options, e.g. Font size, etc.

## Ideas

- Use a hidden folder in the book dir for app data to make reading progress and bookmarks git-manageable. 
