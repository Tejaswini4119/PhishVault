(function() {
  const _0x3bfa = ['form', 'querySelector', 'login-btn', 'addEventListener', 'submit', 'preventDefault', 'textContent', 'Verifying...', 'disabled', 'Login', 'alert', 'Invalid credentials. Please try again.', 'setTimeout'];
  
  const obf = function(index) {
    return _0x3bfa[index];
  };

  document[obf(1)](obf(0))[obf(3)](obf(4), function(event) {
    event[obf(5)]();
    const btn = document[obf(1)]('.' + obf(2));
    btn[obf(6)] = obf(7);
    btn[obf(8)] = true;

    window[obf(12)](() => {
      btn[obf(6)] = obf(9);
      btn[obf(8)] = false;
      window[obf(10)](obf(11));
    }, 2000);
  });
})();
