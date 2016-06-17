# protocoldesign

tornet <files dir> <nodes list>

<files dir> -- absolute path to the directory with peered files
<nodes list> --  a list of nodes in format IP:PORT separated by spaces

tip: put files you need to distribute in the folder _other_ than pft-files (i.e. tornet-files)
     to avoid problems with serving files that start with underscore


# what's left

- test.go -> p2p.go with the rest of the logic

important: don't change the dir structure (127.0.0.1_PORT/pft-files/...)