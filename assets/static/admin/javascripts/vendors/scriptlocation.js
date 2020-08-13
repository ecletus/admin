// Get script file location
// doesn't work for older browsers

function __filename(skip) {
    skip = skip || 0;
    skip++;
    var stack = "stack",
        stackTrace = "stacktrace",
        loc = null;

    try {
        // Invalid code
        0();
    }  catch (ex) {
        var matcher = function (stack, matchedLoc) {
            console.log(matchedLoc);
            return loc = matchedLoc;
        };

        if (stackTrace in ex) { // Opera
            ex[stackTrace].replace(/called from line \d+, column \d+ in (.*):/gm, matcher);
        } else if (stack in ex) {
            let txt = "" + ex[stack],
                s = txt.indexOf('(');
            if (s >= 0) {
                let i;
                for (i = 0; i < skip; i++)
                    s = txt.indexOf('(', s + 3);
                let e = txt.indexOf(')', s + 1);
                loc = txt.substring(s+1, e).replace(/:\d+:\d+$/, '')
            } else if ((s = txt.indexOf('@')) !== -1) {
                let i;
                for (i = 0; i < skip; i++)
                    s = txt.indexOf('@', s + 1);
                let e = txt.indexOf('\n', s + 1);
                loc = txt.substring(s+1, e).replace(/:\d+:\d+$/, '')
            }
        }
    }
    return loc;
}

function __dirname(skip) {
    return path.dir(__filename((skip || 0)+1))
}