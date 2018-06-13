<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Upload file</title>
</head>
<body>
  <form action="/upload/assets" enctype="multipart/form-data" method="POST">
    <input type="file" name="file" />
    <input type="hidden" name="token" value="{{.}}" />
    <input type="submit" value="upload"/>
  </form>
</body>
</html>
