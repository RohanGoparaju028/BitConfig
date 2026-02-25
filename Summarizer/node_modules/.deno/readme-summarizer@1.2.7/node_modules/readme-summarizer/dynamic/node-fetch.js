// we are running in browserify / webpack environment. (dont use fs, child_process, etc..)
const fetch = require('node-fetch');

module.exports.fetch = fetch;
module.exports.env = "node";