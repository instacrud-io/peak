package templates

// AppTmpl : is tmpl to create app
var AppTmpl = `
const express = require("express");
const app = express();
const bodyParser = require("body-parser");
const morgan = require("morgan");
{{range $i, $entity := .Entities}}
const {{$entity}}Routes = require("./controlers/{{$entity}}");
{{end}}

app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());

app.use(morgan("tiny"));


{{range $i, $entity := .Entities}}
		app.use("/{{$entity}}", {{$entity}}Routes);
{{end}}


app.use(function(req, res, next) {
  let err = new Error("Not Found");
  err.status = 404;
  next(err);
});

if (app.get("env") === "development") {
  app.use(function(err, req, res, next) {
    res.status(err.status || 500);
    res.send({
      message: err.message,
      error: err
    });
  });
}

app.listen(3000, function() {
  console.log("Server starting on port 3000!");
});


module.exports = {
  app
};`
