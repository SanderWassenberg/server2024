export function hide(elem, state) {
	elem.hidden = state ?? true;
}

// 'data' can be either a string, object or FormData.
// String is sent as-is, FormData and object are json-serialized.
// Returnvalue may be an exception, use `instanceof Error` to check.
export async function api_post(path, data) {
	const options = {
	    method: 'POST',
	    headers: { 'Content-Type': 'application/json' },
	    body: JSON.stringify((data instanceof FormData) ? Object.fromEntries(data) : data),
	    // you can also send a FormData without stringifying it but them it uses a weird scheme, not json
	};
	try { // fucking exceptions man
		// await here is necessary to catch exceptions that occur inside the promise.
		return await fetch(location.origin + "/api" + path, options);
	} catch(e) {
		return e;
	}
}

// See:
// https://stackoverflow.com/questions/201323/how-can-i-validate-an-email-address-using-a-regular-expression
// https://regexper.com/#(%3F%3A%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B(%3F%3A%5C.%5Ba-z0-9!%23%24%25%26%27*%2B%2F%3D%3F%5E_%60%7B%7C%7D%7E-%5D%2B)*%7C%22(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21%5Cx23-%5Cx5b%5Cx5d-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)*%22)%40(%3F%3A(%3F%3A%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%5C.)%2B%5Ba-z0-9%5D(%3F%3A%5Ba-z0-9-%5D*%5Ba-z0-9%5D)%3F%7C%5C%5B(%3F%3A(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D))%5C.)%7B3%7D(%3F%3A(2(5%5B0-5%5D%7C%5B0-4%5D%5B0-9%5D)%7C1%5B0-9%5D%5B0-9%5D%7C%5B1-9%5D%3F%5B0-9%5D)%7C%5Ba-z0-9-%5D*%5Ba-z0-9%5D%3A(%3F%3A%5B%5Cx01-%5Cx08%5Cx0b%5Cx0c%5Cx0e-%5Cx1f%5Cx21-%5Cx5a%5Cx53-%5Cx7f%5D%7C%5C%5C%5B%5Cx01-%5Cx09%5Cx0b%5Cx0c%5Cx0e-%5Cx7f%5D)%2B)%5C%5D)
// to ensure the whole string is matched and not a portion: https://stackoverflow.com/questions/447250/matching-exact-string-with-javascript
const email_regexp = /^(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])$/;
export function is_valid_email(email) {
	return email_regexp.test(email);
}

// See:
// https://stackoverflow.com/questions/5515869/string-length-in-bytes-in-javascript
export function byte_length(string) {
	return new Blob([string]).size;
}

// returns a function which generates copies of the html in the template string.
//	const generate_html = new_template(`<div>a bunch of html</div>`);
//	elem_a.append(generate_html())
//	elem_b.append(generate_html())
export function new_template(template_string) {
    const template = document.createElement('template');
    template.innerHTML = template_string;

    return () => template.content.cloneNode(true);
}

export function random_in_range(min, max) {
    const diff = max - min;
    return min + Math.random() * diff;
}
export function random_int_in_range(min, max) {
    const diff = max - min;
    return Math.floor(min + Math.random() * diff);
}

export function pick_random(arr_or_string) {
	return arr_or_string[Math.floor(Math.random() * arr_or_string.length)]
}

export function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

