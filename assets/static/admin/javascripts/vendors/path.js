window.path = {
    dir: function (url) {
        url = new URL(url);
        var parts = url.pathname.split("/");
        url.pathname = parts.slice(0, parts.length-1).join("/");
        return url.toString();
    },
    join: function (url, sub) {
        url = new URL(url);
        var parts = url.pathname.split("/");
        for (var i = 1; i < arguments.length; i++) {
            parts[parts.length] = arguments[i]
        }
        url.pathname = parts.join("/");
        return url.toString();
    }
};