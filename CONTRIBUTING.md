# Contributing

## Create a new module

Netm4ul uses several modules. Every module follows an interface and must implements all of the methods to function properly.

We have created one command to help the creation of a module : 
`netm4ul create module --name MyNewModule --short-name mnm --type recon --author edznux`
`--name` defines the structure name and `--short-name` defines the used variable name.
If `--short-name` is not provided, it will be set to the first letter of the `--name`, in lowercase.
You should find your boiler plate code inside the `modules/<type>/<module>` directory (where `<type>` is the `--type` (must be one of `recon`, `report` or `exploit`) and `<module>` is the lowercase `--name`)


### Recon

Recon modules are used to perform recon scan.
- `Run` function will be called on the client side.
- `WriteDb` function will be called on the server side.
- `ParseConfig` is never called outside of the module. It's just required to enforce module to not hardcode value.

### Exploits

//TODO 

### Reports

//TODO

## Create a new database adapter

To add a new database adapter, you will need to implement the models.Database interface (available in the `core/database/models/` directory)

To facilitate the development, you can generate all the boiler-plate with this command :
`netm4ul create adapter --name MyDatabase --short-name md`
`--name` defines the structure name and `--short-name` defines the used variable name.
If `--short-name` is not provided, it will be set to the first letter of the `--name`, in lowercase.
You should find your boiler plate code inside the `core/database/adapters/<mydatabase>` directory (where `<mydatabase>` is the lowercase `--name`)

*Then*, you will need to `Register` your adapter inside `database.go`.

## Create a new load balance algorithm

To add a new algorithm, you will need to implement the algorithm.Algorithm interface (available in the `core/loadbalancing/algorithms/` directory)

To facilitate the development, you can generate all the boiler-plate with this command :
`netm4ul create algorithm --name MyAlgo --short-name ma`
`--name` defines the structure name and `--short-name` defines the used variable name.
If `--short-name` is not provided, it will be set to the first letter of the `--name`, in lowercase.
You should find your boiler plate code inside the `core/loadbalancing/algorithms/<myalgorithm>` directory (where `<myalgorithm>` is the lowercase `--name`)

*Then*, you will need to `Register` your adapter inside `core/loadbalancing/loadbalancing.go`.
