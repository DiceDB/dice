## Install antlr4

Reference: https://github.com/antlr/antlr4/blob/master/doc/getting-started.md

1. Install antlr tools
```bash
pip install antlr4-tools
```

2. Pull the latest jar
```bash
antlr4 
```

## Generate the parser
Run this command from within the parser folder.
```bash
antlr4 -Dlanguage=Go -o ./genfiles DSQL.g4
```
This command will generate the parser files in the genfiles folder, do not modify these files.