export default { test, getExt };
function test(txt) {
    console.log(txt);
}
function getExt(pathName) {
    let pathNameSplit = pathName.split(".");
    // If spl.length is one, it's a visible file with no extension ie. file
    // If spl[0] === "" and spl.length === 2 it's a hidden file with no extension ie. .htaccess
    if (pathNameSplit.length === 1 || (pathNameSplit[0] === "" && pathNameSplit.length === 2)) {
        return "";
    }
    return pathNameSplit.pop().toLowerCase();
}
