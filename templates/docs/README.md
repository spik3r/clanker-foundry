# Documentation building blocks

Copy the smallest block that solves the reader's problem. Replace every
placeholder, then remove optional sections that do not apply.

| Template | Use it when the reader needs to |
| --- | --- |
| [README blocks](readme-blocks.md) | Understand a project and find the right next step |
| [Quickstart](quickstart.md) | Reach and recognise their first successful run |
| [Requirements](requirements.md) | Agree on observable behaviour before implementation |
| [Troubleshooting](troubleshooting.md) | Recover from a specific symptom |
| [Release note](release-note.md) | Know what changed and whether they need to act |

For a new repository, begin with the
[project README starter](../project/project-readme.md) and add only the blocks
you need. Keep requirements in the requirements document, design choices in an
[architecture decision record](../adr/adr-template.md), and runnable steps in
one quickstart. Link to the canonical copy instead of repeating it.

Before publishing, click the links from the destination file, follow the
quickstart in a clean environment, and check that every requirement has a
verification case. Use fictional or sanitised examples.
