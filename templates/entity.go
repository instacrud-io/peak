package templates

// EntityTmpl : is tmpl to create app
var EntityTmpl = `
const express = require("express");
const router = express.Router();
const db = require("../db");

router.get("/", async function(req, res, next) {
  try {
    const results = await db.query("SELECT {{.Fields}} FROM {{ .Entity }} ");
    return res.json(results.rows);
  } catch (err) {
    return next(err);
  }
});

router.post("/", async function(req, res, next) {
  try {
    const result = await db.query(
      "INSERT INTO {{ .Entity }}({{.InsFdName}}) VALUES ({{.InsFields}}) RETURNING *",
      [
          {{ range $key, $value := .InsAttributes }}
		    req.body.{{ $key  }},
	      {{ end }}
      ]
    );
    return res.json(result.rows[0]);
  } catch (err) {
    return next(err);
  }
});

router.get("/:id", async function(req, res, next) {
  try {
    const results = await db.query("SELECT {{.Fields}} FROM {{ .Entity }} WHERE {{.IdName}} = $1",
    [req.params.id]);
    return res.json(results.rows[0]);
  } catch (err) {
    return next(err);
  }
});

router.patch("/:id", async function(req, res, next) {
  try {
    const result = await db.query(
      "UPDATE {{ .Entity }} set {{ .CrFields }} WHERE {{.IdName}} = ${{.UpAttrLen}}",
      [
          {{ range $key, $value := .UpdateAttributes }}
		    req.body.{{ $key  }},
	      {{ end }}
          req.params.id
      ]
    );
    return res.json(result.rows[0]);
  } catch (err) {
    return next(err);
  }
});

router.delete("/:id", async function(req, res, next) {
  try {
    const result = await db.query("DELETE FROM {{ .Entity }} WHERE id = $1", [
      req.params.id
    ]);
    return res.json({ message: "Deleted" });
  } catch (err) {
    return next(err);
  }
});

module.exports = router;`
