package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var (
	fl_listen = flag.String("l", ":8765", "non-tls listen")
	synth     *exec.Cmd
	lang      *exec.Cmd
	input     io.WriteCloser
)

func main() {
	flag.Parse()

	// TODO: start within a container
	synth = exec.Command("scsynth", "-u", "57110")
	lang = exec.Command("sclang", "-u", "57120")

	lang.Stdout = os.Stdout
	synth.Stdout = os.Stdout

	input, _ = lang.StdinPipe()
	input.Write([]byte("s.boot;" + string(0x0c)))

	go synth.Run()
	go startHTTP()

	lang.Start()
	lang.Wait()

}

func startHTTP() {
	http.HandleFunc("/", startHandler)
	http.HandleFunc("/x", execHandler)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	log.Fatal(http.ListenAndServe(*fl_listen, nil))
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
<html>
<head><title>SC</title></head>
<body>
<script>
function play() {
  let req = new XMLHttpRequest();
  let input = document.getElementById("input");
  req.open("POST", "/x");
  req.enctype = "text/plain";
  req.send(input.value);
}
</script>
<h1>SC</h1>
<form method="post" action="/x">
<textarea id="input" name="input" rows="33" cols="80">
(freq: 440).play;
</textarea>
<input type="button" onclick=play()>
</form>
</body>
</html>
`))
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	buf := &bytes.Buffer{}
	io.Copy(buf, io.LimitReader(r.Body, 1024*1024*10))
	log.Printf("Sending %q to sclang\n", buf.String())
	buf.WriteByte(0x0c)
	io.Copy(input, buf)
}
