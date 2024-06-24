import { api_post, hide } from "./util.js";

const row_template  = document.querySelector(".search-table tbody template");

const notif          = document.querySelector("#notif");
const interest_form  = document.querySelector("#interest_form");
const interest_input = document.querySelector("#interest_input");
const search_form    = document.querySelector("#search_form");
const search_table   = document.querySelector(".search-table tbody");
const search_input   = document.querySelector("#search_input");
const search_spinner = document.querySelector("#search_spinner");
const username       = document.querySelector("#username");

(async () => {
	interest_form.addEventListener("submit", set_interest);
	search_form.addEventListener("submit", search);
	let response = await api_post("/user_info", "");
	if (!await check_response(response)) return;

	let info = await response.json();
	document.documentElement.dataset["role"] = info.role;

	interest_input.value = info.interest;
	username.innerText = info.name;

	search();
})();


async function set_interest(e) {
	const response = await api_post("/set_interest", interest_input.value);
	if (!await check_response(response)) return;
	set_notif(`interest updated to ${interest_input.value}`, "green");
}

async function search(e) {
	while (search_table.lastChild) search_table.lastChild.remove();

	hide(search_spinner, false);
	const response = await api_post("/search", search_input.value);
	hide(search_spinner, true);

	if (!await check_response(response)) return;
	let results = await response.json();
	set_notif(`Got ${results.length} result${results.length != 1 ? "s" : ""}`, "black");

	if (results.length < 1) return;

	for (let i = 0; i < results.length; i++) {
		let r = results[i];
		let row = row_template.content.cloneNode(true);

		row.querySelector("[data-field=username]").innerText = r.username;
		row.querySelector("[data-field=interest]").innerText = r.interest;
		let ban_elem = row.querySelector(".ban [data-ban_user]")
		ban_elem.dataset.ban_user = r.username;
		ban_elem.addEventListener("click", click_ban);
		row.querySelector(".chat a").href += r.username;

		search_table.append(row);
	}
}

async function click_ban(e) {
	let who = e.target.dataset.ban_user;
	set_notif(`Banning ${who}...`, "black");
	const response = await api_post("/ban", who);
	if (!await check_response(response)) return;

	set_notif(`Banned ${who}!`, "green");
}

async function check_response(response) {
	if (response instanceof Error) {
		set_notif(`JS Exception: ${response}`, "red");
		return;
	}
	if (!response.ok) {
		const type = response.type != "basic" ? ` (type ${response.type})` : "";
		set_notif(`${response.status} (${response.statusText})${type}: ${await response.text()}`, "red");

		if (response.status === 401 /* unauthorized */) {
			setTimeout(() => {
				location.href = "/login/";
			}, 2000);
		}
		return;
	}

	return true;
}

function set_notif(text, color) {
	notif.classList.remove("wiggle");
	hide(notif, false);
	notif.innerText = text;
	notif.style.color = color;
	if (color === "red") {
		void notif.offsetWidth; // See: https://css-tricks.com/restart-css-animation/
		notif.classList.add("wiggle");
	}
}
