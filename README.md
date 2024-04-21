## forum

### Objectives

This project consists in creating a web forum that allows :

- communication between users.
- associating categories to posts.
- liking and disliking posts and comments.
- filtering posts.

#### Authentication

Client must register and then a login session will be created to access the forum and be able to add posts and comments.
Cookies is used to allow each user to have only one opened session. Each of this sessions contains an expiration date.

Instructions for user registration:
- Enter email
- When the email is already taken it will return an error response.
- Enter username 
- When the username is already taken it will return an error response.
- Enter password
- The password will be encrypted when stored

#### SQLite

For this project the SQLite  database library is used to store accounts, posts, comments and others.

#### Communication

- Only registered users will be able to create posts and comments.
- When registered users are creating a post they can associate one or more categories to it.
- The posts and comments will be visible to all users (registered or not).
- Non-registered users will only be able to see posts and comments.

#### Likes and Dislikes

Only registered users will be able to like or dislike posts and comments.
The number of likes and dislikes are visible to all users (registered or not).

#### Filter

Two types of filters are used:
- Filtering by categories in the home page such as news, technology and lifestyle.
- Filtering by created or liked posts in the profile page.

Note that profile filter is only available for registered users and refer to the logged in user.

### Usage 
```
go run . 
```
open the port

### AUTHORS

- Ticha Nji
- Amir Iqbal


### LICENSES

This program developed within the scope of Reboot.

