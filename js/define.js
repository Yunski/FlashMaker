$(function() {
    define();
});

function define() {
    $("#submit").click(function(e){
        e.preventDefault();
        $("#alert").hide();

        var words = $("#wordbox").val().match(/\S+/g);
        if (words == null || words.length < 2) {
            $("#alert").show();
            $("#alert").text("Need at least two words!");
            return;
        } else if ($("#title").val() == "") {
            $("#alert").show();
            $("#alert").text("Enter a title!");
            return;
        }
        $("#submit").hide();
        $("#loading").show();
        var formData = new FormData();
        formData.append("title", $("#title").val());
        formData.append("content", $("#wordbox").val());
        var file = document.getElementById("file").files[0];
        formData.append("file", file);
        var xhr = new XMLHttpRequest();
        xhr.open("POST", "/define", true);
        xhr.send(formData);
        xhr.addEventListener('readystatechange', function(){
            var resp = xhr.responseText;
            $("#loading").hide();
            $("#resp-url").show();
            $("a[href='']").attr("href", resp)
            $("#resp-link").text("Link to Your Quizlet Set!");
            $("#new-set").show();
        });
    });
}
