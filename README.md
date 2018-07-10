# Introduction

Dman is a project for easy use and security management of database account. It aims to solve the security problem that when developer needs to connect to a mysql database, the dba just give them the username and password for the database directly. It's not convenient, nor it's safe.

Dman tries to hide the username and password when developer needs to connect to a database. Dba use dman server to manage all database account. When a developer need to user an account, he(she) send an application with appkey(an unique name for his(her) program) for the database through dman client. After the dba approve the application, the developer can get the database connection just given appkey and database name. The dman lib will get the account from dman server through a http request.

But, It's not finished completely yet. It does not solve the problem perfectly. The program still can kown the value of username and password. I haven't thought out a method that comletely hide the username and password and just return the connection to the program.

# Installation

There are several steps for dba and developer. The dba need a mysql database to manage all database account. Just run sql in the table.sql file to create tables.

1. Dba need to change the serverAddr and serverPort in the tool.go file and also in the dman.yaml file. The value should be the ip and port which machine the dman server program run. The database config should be 
2. Run "go build" in the project directory.
3. Dba start dman server program using command "./dman server".
4. Run "./install.sh" to build .so file and .h file.
5. Put the .so and .h file to python-binding folder, then you can use the python client to get database connection.

# Usage
The command line help should show infomation for the usage of each command.

More to be updated....

# TODO

1. some better hidding method
2. more bindings

# FAQ

## 1. Why go?
I try to hide the account and just return the connection to the developer. So the best way I can think of is to just give a .sofile to the developer, and return the connection to him, without giving him the source code. 

Also, developer may run their program on different OS, such as Windows, Mac OS, CentOS, Ubuntu, and so on. Golang can compile .so for all these platforms just using one command. No need to build on that platform. It's much convenient for multiple platforms.

## 2. Why not hide the username and password completely?

When calling golang from another language, we need to compile the code to C code, and calling the C code from another language. But the compiled code just support simple data type, not complex struct. So now username and password cannot be hide completely. 

Maybe there is a better idea.

## 3. Main use case?

Many program need to connect to a database. Easy use and security of database account is both valuable for developer and dba.


# Reference

https://medium.com/learning-the-go-programming-language/calling-go-functions-from-other-languages-4c7d8bcc69bf
