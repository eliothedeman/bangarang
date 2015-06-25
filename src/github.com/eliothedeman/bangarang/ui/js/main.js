location.hash = "#";
client = new Client("http://localhost", 8081);
body = document.getElementsByTagName("body").item(0);

b = document.createElement("div")
b.innerHTML = JSON.stringify(client.getServices());

body.appendChild(b);
