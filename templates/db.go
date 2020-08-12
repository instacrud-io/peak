package templates

// DbTmpl : is tmpl to create app
var DbTmpl = `
const { Client } = require("pg");
const client = new Client({
  connectionString: "postgres://oszrwdkweikqbw:2b606c4bd60baf639c557547c1fb0a38f414d774af7c7e00ae12fd6eb064a18d@ec2-34-193-117-204.compute-1.amazonaws.com:5432/d27aq3mo3jlkv6",
  ssl: { rejectUnauthorized: false }
});

client.connect();

module.exports = client;`
