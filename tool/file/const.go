package file

const (
	Ext_TOML  = "toml"
	Ext_CSV   = "csv"
	Ext_PROTO = "proto"
	Ext_GIF   = "gif"
)

//go:generate stringer -type=ReplaceProbe -linecomment

type ReplaceProbe int

const (
	ReplaceProbe_Service ReplaceProbe = iota //api place
	ReplaceProbe_Message                     //message place
)
