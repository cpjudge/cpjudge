# Code-Judge
Online Programming Judge in GO

This application uses [GoBuffalo](https://gobuffalo.io/) framework. 

## Bootstrap 4
Run 
`yarn add bootstrap@4.0.0-beta.2` 
and
`yarn add popper.js` 

## Database

MySQL database is used in this application

Change your mysql username and password in the file `database.yml`. You can also change the databases' name. Currently the database name is `gocoder_dev`

#### To create database 
Run `buffalo db create -a`

#### To create tables in your database
Run `buffalo db migrate up`

Now your database should be ready with the required tables.

## To run the application
Run `buffalo dev run`
