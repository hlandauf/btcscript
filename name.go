package btcscript

import "errors"

// NameScript provides information parsed from a Script. It includes the name
// operation type, the destination address and any operation arguments.
type NameScript struct {
	op      byte
	address *Script
	args    []string
}

var ErrNameEmptyScript = errors.New("pk script contains no opcodes and thus cannot be a valid name script")
var ErrNameOpcodeOutOfRange = errors.New("pk script is not a valid name script because it contains an out-of-range opcode")
var ErrNameNoDropDelimiter = errors.New("pk script is not a valid name script because it does not contain a DROP/2DROP/NOP delimiter")
var ErrNameWrongArgCount = errors.New("pk script is not a valid name script because it does not have the correct number of arguments for the given op type")
var ErrNameUnknownOp = errors.New("pk script is not a valid name script because it has an unknown name op type")

// Attempt to parse a Script in order to find name information.  If the script
// is not a syntactically valid name script, returns an error.
func NewNameScript(s *Script) (*NameScript, error) {
	ns := &NameScript{}

	pkOpcodes := s.scripts[1]

	// Build arguments.

	if len(pkOpcodes) == 0 {
		return nil, ErrNameEmptyScript
	}

	nameOp := pkOpcodes[0].opcode.value

	var i int
	for i = 1; i < len(pkOpcodes); i++ {
		opNum := pkOpcodes[i].opcode.value

		if opNum == OP_DROP || opNum == OP_2DROP || opNum == OP_NOP {
			break
		}

		if opNum < 0 || opNum > OP_PUSHDATA4 {
			return nil, ErrNameOpcodeOutOfRange
		}

		ns.args = append(ns.args, string(pkOpcodes[i].data))
	}

	// Move to after any DROP/NOP opcodes.
	for i = i; i < len(pkOpcodes); i++ {
		opNum := pkOpcodes[i].opcode.value
		if opNum != OP_DROP && opNum != OP_2DROP && opNum != OP_NOP {
			break
		}
	}

	// No DROP/NOP opcodes were encountered before the end of the script, this is
	// invalid.
	if i >= len(pkOpcodes) {
		return nil, ErrNameNoDropDelimiter
	}

	// Check that the name operation type is known and that the right number of
	// arguments are present.
	switch ns.op {
	case OP_NAME_NEW:
		if len(ns.args) != 1 {
			return nil, ErrNameWrongArgCount
		}
	case OP_NAME_FIRSTUPDATE:
		if len(ns.args) != 3 {
			return nil, ErrNameWrongArgCount
		}
	case OP_NAME_UPDATE:
		if len(ns.args) != 2 {
			return nil, ErrNameWrongArgCount
		}
	default:
		return nil, ErrNameUnknownOp
	}

	ns.op = nameOp
	ns.address = s
	return ns, nil
}

// Returns the destination address for the script.
func (ns *NameScript) Address() *Script {
	return ns.address
}

// Returns the name operation type found in the script.
func (ns *NameScript) NameOp() byte {
	return ns.op
}

// Returns true iff the name operation type is FirstUpdate or Update.
func (ns *NameScript) IsAnyUpdate() bool {
	switch ns.op {
	case OP_NAME_NEW:
		return false
	case OP_NAME_FIRSTUPDATE, OP_NAME_UPDATE:
		return true
	default:
		panic("unexpected value in internal field")
	}
}

// Obtains the name name for scripts where IsAnyUpdate() is true.
// Panics otherwise.
func (ns *NameScript) OpName() string {
	switch ns.op {
	case OP_NAME_FIRSTUPDATE, OP_NAME_UPDATE:
		return ns.args[0]
	default:
		panic("called OpName() on non-update name script")
	}
}

// Obtains the name value for scripts where IsAnyUpdate() is true.
// Panics otherwise.
func (ns *NameScript) OpValue() string {
	switch ns.op {
	case OP_NAME_FIRSTUPDATE:
		return ns.args[2]
	case OP_NAME_UPDATE:
		return ns.args[1]
	default:
		panic("called OpValue() on non-update name script")
	}
}

// Returns the random value for FirstUpdate name operations.
// Panics otherwise.
func (ns *NameScript) OpRand() string {
	switch ns.op {
	case OP_NAME_FIRSTUPDATE:
		return ns.args[1]
	default:
		panic("called OpRand() on non-FirstUpdate name script")
	}
}

// Returns the name hash for New name operations.
// Panics otherwise.
func (ns *NameScript) OpHash() string {
	switch ns.op {
	case OP_NAME_NEW:
		return ns.args[0]
	default:
		panic("called OpHash() on non-New name script")
	}
}

// Determines whether a script contains a syntatically valid name script.
func IsNameScript(s *Script) bool {
	_, err := NewNameScript(s)
	return err == nil
}
