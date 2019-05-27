const path = require('path');
const express = require('express');
const proxy = require('express-http-proxy');
const app = express();

const apiURL = process.env.APIURL || 'localhost:8080';

console.log(`API URL is ${apiURL}`);

app.get('/healthcheck', function (req, res) {
    res.send('OK');
});

app.use('/api', proxy(apiURL));
app.use('/', express.static(path.join(__dirname, 'public')));
app.listen(3000);