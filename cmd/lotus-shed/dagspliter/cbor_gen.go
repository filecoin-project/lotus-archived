// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package dagspliter

import (
	"fmt"
	"io"

	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

func (t *Prefix) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{164}); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Version (uint64) (uint64)
	if len("Version") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Version\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Version"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Version")); err != nil {
		return err
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Version)); err != nil {
		return err
	}

	// t.Codec (uint64) (uint64)
	if len("Codec") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Codec\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Codec"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Codec")); err != nil {
		return err
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.Codec)); err != nil {
		return err
	}

	// t.MhType (uint64) (uint64)
	if len("MhType") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"MhType\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("MhType"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("MhType")); err != nil {
		return err
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.MhType)); err != nil {
		return err
	}

	// t.MhLength (int64) (int64)
	if len("MhLength") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"MhLength\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("MhLength"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("MhLength")); err != nil {
		return err
	}

	if t.MhLength >= 0 {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajUnsignedInt, uint64(t.MhLength)); err != nil {
			return err
		}
	} else {
		if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajNegativeInt, uint64(-t.MhLength-1)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Prefix) UnmarshalCBOR(r io.Reader) error {
	*t = Prefix{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("Prefix: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadStringBuf(br, scratch)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t.Version (uint64) (uint64)
		case "Version":

			{

				maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
				if err != nil {
					return err
				}
				if maj != cbg.MajUnsignedInt {
					return fmt.Errorf("wrong type for uint64 field")
				}
				t.Version = uint64(extra)

			}
			// t.Codec (uint64) (uint64)
		case "Codec":

			{

				maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
				if err != nil {
					return err
				}
				if maj != cbg.MajUnsignedInt {
					return fmt.Errorf("wrong type for uint64 field")
				}
				t.Codec = uint64(extra)

			}
			// t.MhType (uint64) (uint64)
		case "MhType":

			{

				maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
				if err != nil {
					return err
				}
				if maj != cbg.MajUnsignedInt {
					return fmt.Errorf("wrong type for uint64 field")
				}
				t.MhType = uint64(extra)

			}
			// t.MhLength (int64) (int64)
		case "MhLength":
			{
				maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
				var extraI int64
				if err != nil {
					return err
				}
				switch maj {
				case cbg.MajUnsignedInt:
					extraI = int64(extra)
					if extraI < 0 {
						return fmt.Errorf("int64 positive overflow")
					}
				case cbg.MajNegativeInt:
					extraI = int64(extra)
					if extraI < 0 {
						return fmt.Errorf("int64 negative oveflow")
					}
					extraI = -1 - extraI
				default:
					return fmt.Errorf("wrong type for int64 field: %d", maj)
				}

				t.MhLength = int64(extraI)
			}

		default:
			return fmt.Errorf("unknown struct field %d: '%s'", i, name)
		}
	}

	return nil
}
func (t *BoxedNode) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{162}); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Prefix (dagspliter.Prefix) (struct)
	if len("Prefix") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Prefix\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Prefix"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Prefix")); err != nil {
		return err
	}

	if err := t.Prefix.MarshalCBOR(w); err != nil {
		return err
	}

	// t.Raw (cid.Cid) (struct)
	if len("Raw") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Raw\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Raw"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Raw")); err != nil {
		return err
	}

	if err := cbg.WriteCidBuf(scratch, w, t.Raw); err != nil {
		return xerrors.Errorf("failed to write cid field t.Raw: %w", err)
	}

	return nil
}

func (t *BoxedNode) UnmarshalCBOR(r io.Reader) error {
	*t = BoxedNode{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("BoxedNode: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadStringBuf(br, scratch)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t.Prefix (dagspliter.Prefix) (struct)
		case "Prefix":

			{

				if err := t.Prefix.UnmarshalCBOR(br); err != nil {
					return xerrors.Errorf("unmarshaling t.Prefix: %w", err)
				}

			}
			// t.Raw (cid.Cid) (struct)
		case "Raw":

			{

				c, err := cbg.ReadCid(br)
				if err != nil {
					return xerrors.Errorf("failed to read cid field t.Raw: %w", err)
				}

				t.Raw = c

			}

		default:
			return fmt.Errorf("unknown struct field %d: '%s'", i, name)
		}
	}

	return nil
}
func (t *Edge) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{162}); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Box (dagspliter.WeakCID) (slice)
	if len("Box") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Box\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Box"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Box")); err != nil {
		return err
	}

	if len(t.Box) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Box was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Box))); err != nil {
		return err
	}

	if _, err := w.Write(t.Box[:]); err != nil {
		return err
	}

	// t.Links (dagspliter.WeakCID) (slice)
	if len("Links") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Links\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Links"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Links")); err != nil {
		return err
	}

	if len(t.Links) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.Links was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajByteString, uint64(len(t.Links))); err != nil {
		return err
	}

	if _, err := w.Write(t.Links[:]); err != nil {
		return err
	}
	return nil
}

func (t *Edge) UnmarshalCBOR(r io.Reader) error {
	*t = Edge{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("Edge: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadStringBuf(br, scratch)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t.Box (dagspliter.WeakCID) (slice)
		case "Box":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Box: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Box = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.Box[:]); err != nil {
				return err
			}
			// t.Links (dagspliter.WeakCID) (slice)
		case "Links":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.ByteArrayMaxLen {
				return fmt.Errorf("t.Links: byte array too large (%d)", extra)
			}
			if maj != cbg.MajByteString {
				return fmt.Errorf("expected byte array")
			}

			if extra > 0 {
				t.Links = make([]uint8, extra)
			}

			if _, err := io.ReadFull(br, t.Links[:]); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown struct field %d: '%s'", i, name)
		}
	}

	return nil
}
func (t *Box) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{163}); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Nodes ([]*dagspliter.BoxedNode) (slice)
	if len("Nodes") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Nodes\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Nodes"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Nodes")); err != nil {
		return err
	}

	if len(t.Nodes) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Nodes was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Nodes))); err != nil {
		return err
	}
	for _, v := range t.Nodes {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}

	// t.Internal ([]cid.Cid) (slice)
	if len("Internal") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Internal\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Internal"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Internal")); err != nil {
		return err
	}

	if len(t.Internal) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Internal was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Internal))); err != nil {
		return err
	}
	for _, v := range t.Internal {
		if err := cbg.WriteCidBuf(scratch, w, v); err != nil {
			return xerrors.Errorf("failed writing cid field t.Internal: %w", err)
		}
	}

	// t.External ([]*dagspliter.Edge) (slice)
	if len("External") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"External\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("External"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("External")); err != nil {
		return err
	}

	if len(t.External) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.External was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.External))); err != nil {
		return err
	}
	for _, v := range t.External {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *Box) UnmarshalCBOR(r io.Reader) error {
	*t = Box{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("Box: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadStringBuf(br, scratch)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t.Nodes ([]*dagspliter.BoxedNode) (slice)
		case "Nodes":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.Nodes: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.Nodes = make([]*BoxedNode, extra)
			}

			for i := 0; i < int(extra); i++ {

				var v BoxedNode
				if err := v.UnmarshalCBOR(br); err != nil {
					return err
				}

				t.Nodes[i] = &v
			}

			// t.Internal ([]cid.Cid) (slice)
		case "Internal":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.Internal: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.Internal = make([]cid.Cid, extra)
			}

			for i := 0; i < int(extra); i++ {

				c, err := cbg.ReadCid(br)
				if err != nil {
					return xerrors.Errorf("reading cid field t.Internal failed: %w", err)
				}
				t.Internal[i] = c
			}

			// t.External ([]*dagspliter.Edge) (slice)
		case "External":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.External: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.External = make([]*Edge, extra)
			}

			for i := 0; i < int(extra); i++ {

				var v Edge
				if err := v.UnmarshalCBOR(br); err != nil {
					return err
				}

				t.External[i] = &v
			}

		default:
			return fmt.Errorf("unknown struct field %d: '%s'", i, name)
		}
	}

	return nil
}
