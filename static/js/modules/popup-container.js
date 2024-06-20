import { new_template } from "./util.js";

const make_body = new_template(`
<slot></slot>
<link rel="stylesheet" href="/css/reset.css">
<link rel="stylesheet" href="/css/style.css">
<style>
:host {
	display: flex;
	flex-direction: row;
	justify-content: center;

	position: fixed;
	bottom: 0;
	left: 0;
	right: 0;
	padding: 5px;
}
</style>
`)


export class PopUpContainer extends HTMLElement {
	static { customElements.define("popup-container", PopUpContainer); }

	constructor() {
		super();

		this.attachShadow({mode: 'open'});
		this.shadowRoot.append(make_body());
	}
}