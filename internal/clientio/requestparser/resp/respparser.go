package respparser

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dicedb/dice/internal/cmd"
)

type RESPType byte

const (
	SimpleString RESPType = '+'
	Error        RESPType = '-'
	Integer      RESPType = ':'
	BulkString   RESPType = '$'
	Array        RESPType = '*'
)

// Common errors
var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnexpectedEOF = errors.New("unexpected EOF")
	ErrProtocolError = errors.New("protocol error")
)

// CRLF is the line delimiter in RESP
var CRLF = []byte{'\r', '\n'}

// Parser is responsible for parsing RESP protocol data
type Parser struct {
	data []byte
	pos  int
}

type Option func(*Parser)

// NewParser creates a new RESP parser
func NewParser() *Parser {
	return &Parser{
		pos: 0,
	}
}

func (p *Parser) SetData(data []byte) {
	p.data = data
}

// Parse parses the entire input and returns a slice of RedisCmd
func (p *Parser) Parse(data []byte) ([]*cmd.RedisCmd, error) {
	p.data = data
	p.pos = 0
	var commands []*cmd.RedisCmd
	for p.pos < len(p.data) {
		c, err := p.parseCommand()
		if err != nil {
			return commands, err
		}

		commands = append(commands, c)
	}

	return commands, nil
}

func (p *Parser) parseCommand() (*cmd.RedisCmd, error) {
	if p.pos >= len(p.data) {
		return nil, ErrUnexpectedEOF
	}

	// A Dice command should always be an array as it follows RESP2 specifications
	elements, err := p.parse()
	if err != nil {
		return nil, fmt.Errorf("error parsing command: %w", err)
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return &cmd.RedisCmd{
		Cmd:  strings.ToUpper(elements[0]),
		Args: elements[1:],
	}, nil
}

func (p *Parser) parse() ([]string, error) {

	count := 1
	if p.data[p.pos] == '*' {
		line, err := p.readLine()
		if err != nil {
			return nil, fmt.Errorf("parse array length: %w", err)
		}

		count, err = strconv.Atoi(string(line[1:])) // Remove '*'
		if err != nil {
			return nil, fmt.Errorf("invalid array length %q: %w", line, err)
		}

		if count < 0 {
			return nil, fmt.Errorf("invalid array length: %d", count)
		}
	}

	result := make([]string, 0, count)
	for i := 0; i < count; i++ {
		val, err := p.ParseOne()
		if err != nil {
			return nil, fmt.Errorf("parse array element %d: %w", i, err)
		}

		if sVal, ok := val.(string); ok {
			result = append(result, sVal)
		} else if ival, ok := val.(int64); ok {
			result = append(result, strconv.FormatInt(ival, 10))
		} else {
			return nil, fmt.Errorf("error during RESP parsing, expected string, got %T", val)
		}
	}

	if len(result) != count {
		return nil, fmt.Errorf("array length mismatch: expected %d, got %d", count, len(result))
	}

	return result, nil
}

func (p *Parser) ParseOne() (any, error) {
	for {
		if p.pos >= len(p.data) {
			return "", ErrUnexpectedEOF
		}

		switch RESPType(p.data[p.pos]) {
		case SimpleString:
			return p.parseSimpleString()
		case Error:
			return p.parseError()
		case Integer:
			return p.parseInteger()
		case BulkString:
			return p.parseBulkString()
		case Array:
			return p.parse()
		default:
			return "", fmt.Errorf("%w: unknown type %c", ErrProtocolError, p.data[p.pos])
		}
	}
}

func (p *Parser) parseSimpleString() (string, error) {
	p.pos++ // Skip the '+'
	line, err := p.readLine()
	if err != nil {
		return "", fmt.Errorf("parse simple string: %w", err)
	}
	return string(line), nil
}

func (p *Parser) parseError() (string, error) {
	p.pos++ // Skip the '-'
	line, err := p.readLine()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}
	return string(line), nil // Preserve the error indicator
}

func (p *Parser) parseInteger() (val int64, err error) {
	p.pos++ // Skip the ':'
	line, err := p.readLine()
	if err != nil {
		return 0, fmt.Errorf("parse integer: %w", err)
	}
	// Validate that it's a valid integer
	val, err = strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", line, err)
	}
	return val, nil
}

func (p *Parser) parseBulkString() (string, error) {
	line, err := p.readLine()
	if err != nil {
		return "", fmt.Errorf("parse bulk string length: %w", err)
	}
	length, err := strconv.Atoi(string(line[1:])) // Remove '$'
	if err != nil {
		return "", fmt.Errorf("invalid bulk string length %q: %w", line, err)
	}

	if length == -1 {
		return "(nil)", nil // Null bulk string
	}

	if length < -1 {
		return "", fmt.Errorf("invalid bulk string length: %d", length)
	}

	if p.pos+length+2 > len(p.data) { // +2 for CRLF
		return "", ErrUnexpectedEOF
	}

	content := p.data[p.pos : p.pos+length]
	p.pos += length + 2 // Skip the string content and CRLF

	// Verify CRLF after content
	if !bytes.Equal(p.data[p.pos-2:p.pos], CRLF) {
		return "", errors.New("bulk string not terminated by CRLF")
	}

	return string(content), nil
}

func (p *Parser) readLine() ([]byte, error) {
	if p.pos >= len(p.data) {
		return nil, ErrUnexpectedEOF
	}

	end := bytes.Index(p.data[p.pos:], CRLF)
	if end == -1 {
		return nil, ErrUnexpectedEOF
	}

	line := p.data[p.pos : p.pos+end]
	p.pos += end + 2 // +2 to move past CRLF
	return line, nil
}
