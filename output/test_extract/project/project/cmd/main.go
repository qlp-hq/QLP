{
  "files": {
    "go.mod": "module simple-server\n\ngo 1.21",
    "main.go": "package main\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n\t\"log\"\n)\n\nfunc main() {\n\thttp.HandleFunc(\"/health\", healthHandler)\n\tlog.Println(\"Server starting on :8080\")\n\tlog.Fatal(http.ListenAndServe(\":8080\", nil))\n}\n\nfunc healthHandler(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"OK\")\n}"
  }
}