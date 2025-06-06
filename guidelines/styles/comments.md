# Adept Comment‑Style Guide

> **Version 0.1 · 2025‑06‑05**
> Applies to all Go source files in the Adept repository.

---

## 1. Header Block (required at top of every Go file)

// <relative/path/to/file.go>
//
// <One‑line package or file synopsis.>
//
// Context
// -------
// (Optional.)  A short paragraph describing why this file exists, its
// relationship to neighbouring code, and any design constraints.
//
// Workflow / Life‑cycle
// ---------------------
//  1. <Step one sentence.>
//  2. <Step two sentence.>
//  3. …
//

**Rules**

* The first line is always the file’s repo‑relative path.  This speeds up search and copy‑paste in reviews.
* A blank `//` separates logical sections.
* Use ASCII hyphens (`-`) in underline headings.
* Prefer numbered workflows for deterministic processes, and bullet lists for unordered notes.

---

## 2. Package Comments

* Package comment lives in the file that defines `package <name>` if no other file has the synopsis.
* Begin with the package name, e.g., `// Package tenant handles lazy‑loaded site instances.`
* Follow with a blank line, then longer narrative as needed.

---

## 3. Exported Identifiers

Standard GoDoc rules apply *plus* the Adept prose style:

// Load pre‑warms the tenant cache for a slice of hostnames.
//
// It returns the count of hosts successfully primed and a summary error.
func Load(ctx context.Context, hosts []string) (int, error) { … }

* Start with the identifier name.
* Follow with a single‑sentence purpose line.
* Optional paragraphs expand on behaviour and edge cases.

---

## 4. Inline Comments

* Keep inline `//` comments terse.  Favour nouns over verbs (`// mutex protects stats cache`).
* Use TODO tags with your initials and a date, e.g., `// TODO(BJY 2025‑07‑01): eliminate copy.`

---

## 5. Change Workflow When Editing Files

1. **Fetch original file** (commit or working copy).
2. Apply edits.
3. **Preserve** existing comments and update step numbers or bullets to stay accurate.
4. Output the *full* file after modifications—never a patch or snippet—so reviewers see the complete context.
5. Ensure `go vet ./...` passes before committing.

---

## 6. Quick Reference

| Area                 | Rule                                                            |
| -------------------- | --------------------------------------------------------------- |
| Oxford comma         | Always.                                                         |
| Sentence spacing     | Two spaces after every period.                                  |
| m‑dash               | Do **not** use.  Replace with colon, semicolon, or parentheses. |
| Max line length      | 100 columns.                                                    |
| Headings             | ASCII hyphens (`‑`) for underline; no equals signs.             |
| Bullet indent        | One space after `//`, then `•`.                                 |
| Numbered list indent | Align under leading digit.                                      |

---

*End of guide.  Update version header if substantial changes occur.*
