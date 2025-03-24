# Development Setup for Docs

We use [Astro](https://astro.build/) framework to power the [dicedb.io website](https://dicedb.io) and [Starlight](https://starlight.astro.build/) to power the docs. Once you have NodeJS installed, fire the following commands to get your local version of [dicedb.io](https://dicedb.io) running.

```bash
$ cd docs
$ npm install
$ npm run dev
```

Once the server starts, visit http://localhost:4321/ in your favourite browser. This runs with a hot reload which means any changes you make in the website and the documentation can be instantly viewed on the browser.

### Docs directory structure

1. `docs/src/content/docs/commands` is where all the commands are documented
2. `docs/src/content/docs/tutorials` is where all the tutorials are documented
