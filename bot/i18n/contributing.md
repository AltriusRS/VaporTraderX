# Contributing

## Adding a new language

To add a new language, create a new JSON file in the `i18n` directory with the name of the language's ISO code (e.g. `en-US.json5`). Then, simply fill out the fields in the file with the translated values.
Please note that a value that looks like ``$%NAME%$`` is replaced with a value, and so translations should try to work around this limitation.

## Format
This project uses [JSON5](https://json5.org/) for its translation files. This allows for comments, and is much more readable than the standard JSON specification.

### Examples

For a functional example, see the `en-US.json5` file in this directory, as it is the default language file, and is considered complete.

## Supported Languages
- Any language, if you have the time (and the motivation) to translate it.

| Language | Status | Maintainers |
|----------|--------|-------------|
| English  | ✅     | Altrius     |