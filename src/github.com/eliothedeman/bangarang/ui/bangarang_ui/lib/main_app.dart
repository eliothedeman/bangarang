// Copyright (c) 2015, <your name>. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

import 'dart:html';

import 'package:paper_elements/paper_input.dart';
import 'package:paper_elements/paper_button.dart';
import 'package:polymer/polymer.dart';

/// A Polymer `<main-app>` element.
@CustomTag('main-app')
class MainApp extends PolymerElement {
	String username = "";
	String password = "";

	void setUsername(Event event, Object object, PaperInput target) {
		username = target.value;
	}

	void setPassword(Event event, Object object, PaperInput target) {
		password = target.value;
	}

	void login(Event event, Object object) {
		print("$username $password");
	}


  	/// Constructor used to create instance of MainApp.
  	MainApp.created() : super.created();
}



