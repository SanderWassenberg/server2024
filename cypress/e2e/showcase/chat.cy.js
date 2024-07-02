

describe('chat', () => {
  let aoeu_cookie, asdf_cookie;
  
  before(() => {
    cy.login("asdf", "asdf", (cookie) => {
      asdf_cookie = cookie;
    }).login("aoeu", "aoeu", (cookie) => {
      aoeu_cookie = cookie
    });
  });


  it("should remember messages that were sent", ()=>{
    // send message
    cy.change_login(aoeu_cookie);
    cy.visit("/chat/?with=asdf");
    const message_from_aoeu = makeid(10);
    cy.get("#textbox").type(message_from_aoeu + "{enter}");
    cy.get(".msg.self:last-child").should("have.text", message_from_aoeu);

    // receive message
    cy.change_login(asdf_cookie);
    cy.visit("/chat/?with=aoeu");
    cy.get(".msg:not(.self):last-child").should("have.text", message_from_aoeu);

    // send message back
    const message_from_asdf = makeid(10);
    cy.get("#textbox").type(message_from_asdf + "{enter}");
    cy.get(".msg.self:last-child").should("have.text", message_from_asdf);

    // receive message back
    cy.change_login(aoeu_cookie);
    cy.visit("/chat/?with=asdf");
    cy.get(".msg:not(.self):last-child").should("have.text", message_from_asdf);
  });
});

function makeid(length) {
  let result = [];
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  for (let i = 0; i < length; i++) {
    result.push(characters.charAt(Math.floor(Math.random() * charactersLength)));
  }
  return result.join("");
}