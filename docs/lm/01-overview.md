# Overview

Sources: `index.html`, `grammar.html`, `unrestricted.html`, `paradigm_shift.html`, `lex_and_yacc_and_all_that.html`, `faq.html`, `project.html`.

## What it is

The Language Machine is "a toolkit for language and grammar". At its core is an engine that applies **grammatical substitution rules**: rules that recognise strings, which may contain grammatical symbols, and rewrite them. The engine works **online, one symbol at a time**. Applying a rule means a **recognition phase** followed by a **substitution phase**:

- Either phase may produce any number of symbols, including none.
- The phases of one rule application usually occur **nested inside** the phases of other rule applications.

Rules are written in **lm2n** (language meta notation, or language machine notation). Side-effect actions are written in a notation that is essentially a subset of JavaScript. Heavy computation is meant to go out to external C or D procedures.

The engine does not build a parse tree or derivation by itself. The result of an analysis lies entirely in the **variable bindings** that rules create explicitly. Those bindings follow scope rules that reflect the structure of the analysis, but they do not have to encode it.

## Main features (from the home page)

- Rules describe how to recognise and transform grammatical input. The left side of a rule is a pattern, and the right side says how the pattern is treated.
- Both sides are unrestricted pattern generators.
- It supports multiple named grammars, rule priorities, left recursion, right recursion and central (bracketing) recursion.
- It has variables and associative arrays, using a subset of JavaScript.
- A transformed representation can contain actions and side effects, and can itself be analysed again as input.
- It can run as a free-standing engine or be used as a shared library, and rules can be packaged together with the engine.
- It has a simple interface to external procedures in C and D.
- It has built-in diagnostics, including the **lm-diagram** generator.
- Its metalanguage compilers are written in lm2n itself and share one frontend. Compiled rules can be wrapped as shell scripts or as C or D programs, or compiled to C or D code.
- lm2n source can be treated as wiki text in a subset of MediaWiki markup, so the source doubles as literate documentation.

## Styles of analysis

The engine can combine left-recursive, right-recursive and centrally recursive (bracketing) rules. A rule can be triggered in three ways:

- by the **current context and the current input** together. This is the most efficient kind.
- by the **input regardless of context**. This gives bottom-up, LR(k)-style analysis.
- by the **context regardless of input**. This gives goal-seeking, top-down, LL(k)-style recursive descent.

A rule with an empty substitution phase deletes or ignores what it recognises. A rule with an empty recognition phase produces material "out of thin air", which is typically used for defaults.

## The theory: analytic, not generative

Chomsky's 1957 account of a grammar has four parts:

- terminal symbols
- nonterminal symbols
- a finite set of rewriting rules
- a preferred start symbol

His hierarchy runs from type 0 (unrestricted) through type 1 (context-sensitive) and type 2 (context-free, tree-shaped) to type 3 (regular).

Most later work kept the **generative** view, in which the grammar generates sentences, and so concentrated on context-free and regular grammars. The Language Machine takes the **analytic** view instead: rules go *from* the sentences *towards* the grammar.

```
generative:  sequence-of-symbols-in-the-grammar          => sequence-nearer-the-sentences
analytic:    sequence-nearer-the-sentences-of-the-language => sequence-of-symbols-in-the-grammar
lm2n:         pattern-to-recognise <- replacement ;
```

The site makes these claims:

- Unrestricted *generative* grammars are hard to understand and to apply. Unrestricted *analytic* grammars are relatively easy to understand, and it is possible, though not easy, to build efficient engines that apply them directly.
- *Analysis precedes generation.* To choose which generative rule to apply you already have to recognise a pattern.
- An analytic system that can substitute more symbols than it recognises can also act generatively. The reverse is not true.
- The engine contains the lambda calculus (see `lambda.html`), so it is universal. That is consistent with it implementing type-0 grammars.
- A rule behaves like a **substitution macro** whose pattern and replacement can both contain grammatical symbols, and whose "parameters" are not tied to a fixed format.

### Why not BNF?

To write BNF in lm2n, you split each alternative into its own rule and write each rule the other way round:

```
this_or_that -> 'this' | 'that';      // BNF

'this' <- this_or_that;               // lm2n
'that' <- this_or_that;
'this' <- this_or_that :"this";       // ... then add "parameters"
'that' <- this_or_that :"that";
- this_or_that :Selector noun :Selected
      <- selection :{ 'you chose ' Selector ' ' Selected '?' };
```

### Why not lex and yacc?

The traditional toolchain has several parts, each with its own notation: token definitions, a lex/flex scanner, a yacc/bison grammar, semantic actions in a host language, a driver program and a build system. The Language Machine uses one notation for everything from lexical detail through overall structure to actions. It does not aim to be an optimising compiler generator. It aims to be efficient enough, and quick to write and maintain.

## History

- **1975:** David Hendry and Peri Hankey began developing Hendry's insight into Chomsky's theory.
- **Earlier implementations:** Hankey wrote versions in assembler and then C. They were used for a C translator, several FORTRAN front ends, COBOL dialect translators and a 4GL report-language translator, and they stayed in use for years.
- **A recent predecessor:** an OCaml version.
- **2005–2006 version:** a shared library written in D (built with gdc), with a minimal `lm2` main program and the lm2n compilers. The site states that the full build and test of the compilers takes about 3 minutes from a minimal bootstrap to completion.
- **Project resources:** a SourceForge project page, a dsource.org forum and SVN repository, and discussion on Lambda the Ultimate.
- **Language Machine 2 (this repository):** a Go reimplementation by Mikhail Sorochan, licensed GPLv3. The original was GPLv2.
