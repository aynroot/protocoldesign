# protocoldesign

tornet <files dir> <nodes list>

<files dir> -- absolute path to the directory with peered files
<nodes list> --  a list of nodes in format IP:PORT separated by spaces


# how to run

in addition to tornet also run 
    test.go PORT
where port is one of the specified in tornet node parameters
it will listen to the distributed chunks from the tornet

# what's left

1) tornet file

2) test.go -> p2p.go with the rest of the logic

important: don't change the dir structure (127.0.0.1_PORT/pft-files/...)