🌐 MeshBase: P2P Service Mesh & Tunneling Network
MeshBase is a decentralized peer-to-peer (P2P) networking tool built in Go that allows Docker containers to be discoverable and reachable across any network without public IPs or complex port forwarding. By leveraging a Distributed Hash Table (DHT) and libp2p stream tunneling, MeshBase creates a global "virtual network" for your services.
MeshBase — a peer-to-peer (P2P) mesh network for deploying services securely across nodes.
This MVP allows a single containerized service to run on a VPS, and users can connect to it via a custom subdomain with HTTPS.

Features
+ P2P mesh networking powered by libp2p
+ Capability-based peer discovery (RAM, CPU, GPU, Disk)
+ Live resource monitoring between nodes
+ Single container deployment for MVP
+ Access services via *.meshbase.online wildcard domain
+ HTTPS secured using Let’s Encrypt (via Caddy reverse proxy)

📦 Installation & Build
1. Make it Executable
    - Give the script permission to run:

      chmod +x nodemesh.sh

2. Usage
Now you can start your node with one simple command:
      - To run with a different port (e.g., 5000):
     
      ./nodemesh.sh 5000

