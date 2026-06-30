# Contributing to BFG: Encore

BFG: Encore is a small project so the process is pretty simple, just open an issue or pull request on [GitHub](https://github.com/Paficent/BFGEncore). There is also a [Codeberg](https://codeberg.org/Paficent/jeode) mirror, but its easiest to have PR's sent via GitHub for the time being.

## Code Style

Ideally contributors will use [gopls](https://github.com/golang/tools/tree/master/gopls) as a language server for development.
Most IDEs already integrate with the language server so this should not be an issue.

On top of that, I ask that you use gopls' formatting features on all modified files to keep a consistent code style.

## Generative AI
This project is too small to ban the usage of generative AI flat out, however I require everyone using it to write an acknowledgment of its usage. This is because generative AI is not perfect, causes hallucinations, and tends to take the path of least resistance leading to worse code.

For the purposes of this project, generative AI usage is when AI writes more than 33% of the code you are submitting. AI based code completion tools do not need to be acknowledged. 

## Pull Requests

- Keep changes focused (one fix or feature per PR)
- Test all modified features in the game before submitting
- Describe what your change does and why
- Include an AI acknowledgment if necessary
