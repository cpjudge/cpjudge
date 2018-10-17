# CP-Judge
Online Programming Judge in GO

This application uses [GoBuffalo](https://gobuffalo.io/) framework. 

## Installing Dependencies
Use `yarn install` to install all the required packages

## Database

The application uses MySQL database. Configure the mysql user and database in `database.yml`.

#### To create databases 
Run `buffalo db create -a`

#### To create tables in your database
Run `buffalo db migrate up`

## To run the application
Run `buffalo dev run`
