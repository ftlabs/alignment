var Meditation = (function() {

	var haikuData;

	function urlParam(name){
	    var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
	    if (results==null){
	       return null;
	    }
	    else{
	       return results[1] || 0;
	    }
	}

	function getAndProcessJsonThen( thenFn ) {
		var jsonUrl = "/data/meditation_haiku.json";
		var oReq    = new XMLHttpRequest();
		oReq.onload = processJson;
		oReq.open("get", jsonUrl, true);
		oReq.send();

		function processJson(e) {
			if (this.status == 200) {
				haikuData = JSON.parse(this.responseText);
				// process the json
				//...
			}
			thenFn();
		}
	}

	function displayHaiku() {
		var haikuId = urlParam('haiku');
		var theme   = urlParam('theme') || "DATE";

		// locate haiku
		// - have fallback if not found
		// construct card
		// - image
		// - text
		// - author
		// - buttons
		// inject into page 
		
		var textElt = document.getElementsByClassName("haiku-text")[0];
		textElt.innerHTML = "this is not a haiku";
		textElt.innerHTML = haikuData[0]['TextWithBreaks'];

		// var title           = urlParam('title') || "Carousel";
		// var titleElement    = $('.title');
		// titleElement.html(decodeURIComponent(title).replace(/"/g,""));
		// var footerText      = urlParam('footer') || "Financial Times";
		// var footerElement   = $('.footer-text');
		// footerElement.html(decodeURIComponent(footerText).replace(/"/g,""));
		// if (this.status == 200) {
		//     var data            = JSON.parse(this.responseText);
		//     var carouselElement = $('.carousel');
		//     var htmlList        = data.map(function(it){
		//     	return "\n" + '<div class="item"><div class="text">' + it['text'].replace(/\n/g, '<BR>') + '</div></div>';
		//     });
		//     carouselElement.html(htmlList);
		// }
	}

	return {
		getAndProcessJsonThen:  getAndProcessJsonThen,
		displayHaiku: 			displayHaiku
	};

})();

Meditation.getAndProcessJsonThen( Meditation.displayHaiku );