#Readme for AggreGATOR, a RSS-Feed database for CMD use
1. To use, the user needs GO and Postgres installed, to run the code and database respectively.
2. Using the CLI, run "go install github.com/AleksZieba/gator@latest" to install gator. Gator/AggreGATOR is a CLI tool to read and modify Feeds and databases.
3. In the root of your Repository, create a .gatorconfig file with "{ "db_url": "postgresql://USERNAME:PASSWORD@localhost:5432/DBNAME?sslmode=disable", "current_user_name": "YOUR_USER_NAME" }"
  replace Username, password, dbname, current_user_name and your_user_name with the applicable values of your host.
4.With all of this in place you can use any CLI to check out Feeds, so long as you have the URL for it.
Some valid commands are :
  - reset, to reset the database to its empty state
  - register <username>, to register a user
  - login <username>, to login as a registered user
  - users, to see which users are registered, and which one is logged in
  - addfeed <URL>, to add a URL to get feeds from
  - agg <time>, to periodically check for all feeds from the given url, every 'time'
  - feeds, to see all available feeds
  - follow <feed>, add the given feed to the logged in users feeds to follow
  - unfollow <feed>, remove the given feed from the logged in users feeds to follow
  - following, show all feeds the current logged in user is following
  - browse <x>, show x feeds that have not been seen yet, or not for a long time.
