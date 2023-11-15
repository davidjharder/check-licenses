# `check-lecenses` utility

## Purpose

This utility checks the contents of the `license` key in `package.yml` files against a list of [SPDX indentifiers](https://spdx.org/licenses/) and prints the location of `package.yml` with bad matches.

## Usage

The utility expects to run from the root of the `packages` monorepo.

If a `licenses.json` file does not already exist in the `common/GO/check-licenses` directory, a new copy is wget'ed from the SPDX domain.

## Limitations

- The utility does not understand comments: `MIT # Comment here` is flagged as a bad match, even though `MIT` is a valid identifier.
- The utility does not understand SPDX License Exceptions: `LGPL-2.0-or-later WITH WxWindows-exception-3.1`
- Some multiline `license` keys have a pipe: `license:   |` this causes the subsequent identifiers to be parsed as a single string with line breaks, and is therefore a flagged as a bad match. The pipe should be removed.
- The `slices` package used was only introduced in go 1.18. To make this package compatible with go 1.18 the package would have to be imported from the experimental library, it is in the standard library as of 1.21.

