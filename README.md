# protocoldesign


# testbed

prepare -p <port1>
prepare -p <port2>
tornet -p <port> <files dir> <nodes list>

<port> -- on which port to bind the tracker
<port1>, <port2> -- on whic ports to bind the p2p nodes that accepts the chunks
<files dir> -- absolute path to the directory with peered files
<nodes list> --  a list of nodes in format IP:PORT separated by spaces (use nodes which run 'prepare')

tip: put files you need to distribute in the folder _other_ than pft-files (i.e. tornet-files)
     to avoid problems with serving files that start with underscore


after tornet terminates, you will have chunks on each node that have run 'prepare' + torrent files generated (torrent-files/*)
stop 'prepare' process on nodes


# tornet

either don't stop the tornet after the preparation, or start it again without any parameters (it will load the state from tornet.json)

run two nodes that store the chunks
and one more node that downloads the specific file
for example:
    p2p_node.go -p 4466
    p2p_node.go -p 4467
    p2p_node.go -p 4468 torrent-files/test.pdf.torrent
    
    
current state of the tornet tracker is available in tornet.json file (stores all served files, their chunks and corresponding nodes)