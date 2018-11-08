# Roadmap

## Dev

- [] CLI
    - [] UI
        - [] Nodes graphs
        - [] IPs graphs
        - [] Domains graphs
        - [] Ports graphs
        - [] URI graphs
    - [] Setup 
        - [] Ask and install the WebUI
- [] Core
    - [] API
        - [] Status of the application (API, Server, WebUI, Users ...)
        - [] CORS : load from config file
        - [] GetRoutesByIP
        - [] DeleteProject
        - [] GetRawsByProject
        - [] GetRawsByModule
    - [] Client
        - [] Create UUID for each nodes and save it. (consistency across reboot, disconnect etc...)
        - [] Get real networking capacity (not hardcoded)
    - [] Database
        - [] JsonDB 
            - [] everything
        - [] MongoDB 
            - [] everything
        - [] PostgreSQL
            - [] Setting up authentification
            - [] Correctly implements Create* functions
                - [] IP
                - [] Domain
                - [] Port
                - [] URI
            - [] Correctly implements Update* functions
                - [] IP
                - [] Domain
                - [] Port
                - [] URI
            - [] Correctly implements Delete* functions
                - [] IP
                - [] Domain
                - [] Port
                - [] URI
        - [] TestAdapter
            - [] Implements Delete* functions
                - [] IP
                - [] Domain
                - [] Port
                - [] URI
- [] Modules
    - [] Nmap
    - [] Masscan
    - [] Web page scrapper (retrieve links in a webpage)
    - [] dns bruteforcing
    - [] URL bruteforcing
    - [] Shodan


## Documentation