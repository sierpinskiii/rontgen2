window.onload = function () {
    var slideimg = document.getElementById("slideimg"); 
    var slidequizz = document.getElementById("slidequizz");
    var screen = document.getElementById("screen");

    function updateScreen(id, flag, html) {
        if (id == "0") {
            document.getElementById(id).innerHTML = html;
        }
    }

    // setInterval(updateScreen, 1000);


    // Long Poll
    async function subscribe() {
      let response = await fetch("/screen");

      if (response.status == 502) {
        // Status 502 is a connection timeout error,
        // may happen when the connection was pending for too long,
        // and the remote server or a proxy closed it
        // let's reconnect
        await subscribe();
      } else if (response.status != 200) {
        // An error - let's show it
        console.log(response.statusText);
        // Reconnect in one second
        await new Promise(resolve => setTimeout(resolve, 1000));
        await subscribe();
      } else {
        // Get and show the message
        let message = await response.text();
        console.log(message);
        // updateScreen(message);
        document.getElementById("screen").innerHTML = message;
        // Call subscribe() again to get the next message
        await subscribe();
      }
    }

    subscribe();

}

var slidesize = 100;

function smallerSlide() {
    // var element = document.getElementById("slideimg");
    var element = document.getElementById("screen");
    slidesize -= 5;
    element.style.width = slidesize + "%";
}
    
function largerSlide() {
    // var element = document.getElementById("slideimg");
    var element = document.getElementById("screen");
    slidesize += 5;
    element.style.width = slidesize + "%";
}
