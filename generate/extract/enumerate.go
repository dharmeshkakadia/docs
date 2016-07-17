package extract

import (
	"fmt"
	"math/rand"
)

func (g Grammar) Enumerate(prods Sequences) Sequences {
	var res Sequences
	for _, seq := range prods {
		seqs := g.EnumerateSequenceOld(seq)
		for _, s := range seqs {
			res = append(res, s)
		}
	}
	return res
}

func (g Grammar) EnumerateSequence(s Sequence) Sequence {
	var opts []Sequences
	for _, e := range s {
		switch e := e.(type) {
		case Literal:
			opts = append(opts, Sequences{Sequence{e}})
		case Sequence:
			opts = append(opts, Sequences{e})
		case Repeat:
			opts = append(opts, Sequences{Sequence{e.Expression}})
		case Token:
			switch e {
			case
				//"col_name_keyword",
				//"qualified_name",
				//"unreserved_keyword",
				//"name",
				//"reserved_keyword",
				"":
				opts = append(opts, Sequences{Sequence{Literal("name")}})
				continue
			case "a_expr", "b_expr", "c_expr":
				opts = append(opts, Sequences{Sequence{Literal("TRUE")}})
				continue
			}
			prods := g[string(e)]
			var next Sequences
			for _, seq := range prods {
				next = append(next, seq.(Sequence))
			}
			opts = append(opts, next)
		default:
			panic(fmt.Errorf("unknown expr type: %T: %v", e, e))
		}
	}
	var seq Sequence
	for _, opt := range opts {
		s := opt[rand.Intn(len(opt))]
		seq = append(seq, s...)
	}
	return seq
}

func (g Grammar) EnumerateSequenceOld(s Sequence) Sequences {
	var seqs Sequences
	for _, e := range s {
		switch e := e.(type) {
		case Literal, Sequence:
			seqs = seqs.Append(e)
		case Repeat:
			seqs = seqs.Append(e.Expression)
		case Token:
			switch e {
			case
				"col_name_keyword",
				"qualified_name",
				"unreserved_keyword",
				"name",
				//"a_expr",
				"b_expr",
				"c_expr",
				"":
				seqs = seqs.Append(e)
				continue
			}
			/*
				if !allTokens[e] {
					fmt.Println("TOKEN", e)
					allTokens[e] = true
				}
				//*/
			prods := g[string(e)]
			var next Sequences
			for _, seq := range prods {
				b := seqs.Append(seq)
				next = append(next, b...)
			}
			seqs = next
		default:
			panic(fmt.Errorf("unknown expr type: %T: %v", e, e))
		}
	}
	return seqs
}

func (seqs Sequences) Append(e Expression) Sequences {
	seq, ok := e.(Sequence)
	if !ok {
		seq = Sequence{e}
	}
	if len(seqs) == 0 {
		return Sequences{seq}
	}
	var ret Sequences
	for _, s := range seqs {
		n := append(Sequence{}, s...)
		n = append(n, seq...)
		ret = append(ret, n)
	}
	return ret
}

type Sequences []Sequence

func (seqs Sequences) String() string {
	var prods Productions
	for _, s := range seqs {
		prods = append(prods, s)
	}
	return prods.String()
}
