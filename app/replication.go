package main

type Role int

const (
	master Role = iota
	slave
)

type Replication struct {
	role                           Role
	connected_slaves               int
	master_replid                  string
	master_repl_offset             int
	second_repl_offset             int
	repl_backlog_active            int
	repl_backlog_size              int
	repl_backlog_first_byte_offset int
	repl_backlog_histlen           int
}
