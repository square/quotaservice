### Quotaservice UI

![quotaservice](images/quotaservice.gif)
![quotaservice-stats](images/quotaservice-stats.gif)

#### How?

For development, as long as your QuotaService is running the admin API at port 8080, you can use the following.

```
$ nvm use
Now using node v10.13.0 (npm v6.4.1)
$ npm install
...
$ npm run start
```

Visiting localhost:3000 will enable you to develop the UI application with code reloading.

For a production, minified version:

```
$ npm run dist
```

For code linting:

```
$ npm run lint
```

For testing:

```
$ npm run test
```
