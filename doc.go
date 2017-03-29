/*
Package arc provides tools and packages which allow working with ARC assembly
source code.

package ast implements the ARC abstract syntax tree.
package build implements the ARC assembler.
package fmt implements the ARC source code formater.
package internal implements generic helper functions shared between multiple
packages.
package parser implements the ARC parser.
package scanner implements the ARC tokenizer.
package simulator implements an ARC simulator.
package token implements ARC lexical tokens and token positions.
package vet implements the ARC source code validator.

The arc tool provides a command line interface to these packages and is located
in cmd/arc.
*/
package arc
