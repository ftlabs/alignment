var Meditation = (function() {

	var haikuData;
	var haikuById         = {}; // all haiku indexed by their id
	var haikuListsByTheme = {}; // a list of haiku id for each theme
	var coreThemes        = []; // a list of themes
	var okAuthorsHash     = {}; // a hash of authors with more than one haiku
	var numHaiku;
	var defaultHaiku = 1;
	var defaultTheme = 'IMAGERY';

	function urlParam(name){
	    var results = new RegExp('[\?&]' + name + '=([^&#]*)').exec(window.location.href);
	    if (results==null){
	       return null;
	    }
	    else{
	       return results[1] || 0;
	    }
	}

	// Returns a random integer between min (included) and max (included)
	// Using Math.round() will give you a non-uniform distribution!
	function getRandomIntInclusive(min, max) {
	  min = Math.ceil(min);
	  max = Math.floor(max);
	  return Math.floor(Math.random() * (max - min + 1)) + min;
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
				
				var count = 0;
				var knownCoreThemes = {};
				var knownAuthors    = {};

				haikuData.forEach(function(haiku){
					count = count + 1;
					var id = count; // for now, the haiku id is the index of it in the input data
					haikuById[id] = haiku;

					haiku['ProminentColoursByName'] = {};
					haiku['ProminentColours'].forEach(function(pc){
						haiku['ProminentColoursByName'][pc['Name']] = pc;
					});

					haiku['Themes'].forEach(function(theme){
						if (! (theme in knownCoreThemes)) {
							haikuListsByTheme[theme] = [];
							knownCoreThemes[theme] = true;
						};
						haikuListsByTheme[theme].push(id);

						var author = haiku['Author'];
						if (! (author in haikuListsByTheme)) {
							haikuListsByTheme[author] = [];
							knownAuthors[author] = true;
						};
						haikuListsByTheme[author].push(id);
					});

					coreThemes = Object.keys(knownCoreThemes);
				});

				Object.keys(knownAuthors).forEach(function(author){
					if (haikuListsByTheme[author].length > 1) {
						okAuthorsHash[author] = true;
					};
				});

				numHaiku = haikuData.length;
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
		// -- create a Next/Previous button, and the remaining theme buttons
		// inject into page 
		var id = getRandomIntInclusive(1,numHaiku);
		var haiku = haikuById[id];

		var cardElt = document.getElementsByClassName("haiku-card")[0];
		var prominentColor = haiku['ProminentColours'][0];
		if ('LightMuted' in haiku['ProminentColoursByName']) {
			prominentColor = haiku['ProminentColoursByName']['LightMuted'];
		};
		cardElt.style.backgroundColor = prominentColor['RGBHex'];

		var textElt = document.getElementsByClassName("haiku-text")[0];
		textElt.innerHTML = haiku['TextWithBreaks'];

		var imgElt = document.getElementsByClassName("haiku-image")[0];
		imgElt.src = haiku['ImageUrl'];

		var authorElt = document.getElementsByClassName("haiku-author")[0];
		authorElt.innerHTML = haiku['Author'];

		var navElt = document.getElementsByClassName("haiku-nav")[0];
		navElt.innerHTML = haiku['Themes'].join(', ');
	}

	return {
		getAndProcessJsonThen:  getAndProcessJsonThen,
		displayHaiku: 			displayHaiku
	};

})();

Meditation.getAndProcessJsonThen( Meditation.displayHaiku );