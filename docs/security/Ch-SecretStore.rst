###################
Secret Store
###################

There are all kinds of secrets used within EdgeX micro services, such as tokens, passwords, certificates etc. The secret store serves as the central repository to keep these secrets. The developers of other EdgeX micro services utilize the secret store to create, store and retrieve  secrets relevant to their micro service. The communications between secret store and other micro services are encrypted to prevent man-in-the-middle attacks. 
Currently the EdgeX secret store is implemented with HashiCorp Vault, a 3rd party open source tool for managing secrets. 

