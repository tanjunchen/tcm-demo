const Koa = require('koa');
const path = require('path');
require('koa-router');
const static = require('koa-static');
const bodyParser = require('koa-bodyparser');

const app = new Koa();
const staticPath = './static/dist';

app.use(bodyParser());

app.use(static(path.join(__dirname, staticPath)));

const concurrency = process.env.CONCURRENCY, ip = process.env.INSTANCE_IP;

if (concurrency) {
    app.use(require('./router/index-v2').routes());
} else {
    app.use(require('./router/index-v1').routes());
}

if (ip) {
    app.listen(7000, ip);
    console.log('start success, port: 7000, ip ', ip);
} else {
    app.listen(7000);
    console.log('start success, port: 7000');
}
